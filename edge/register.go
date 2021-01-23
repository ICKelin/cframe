package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/edge/vpc"

	log "github.com/ICKelin/cframe/pkg/logs"
)

type Registry struct {
	srv    string
	secret string
	server *Server

	//heart beat channel
	hbchan chan struct{}

	// report channel
	reportchan chan struct{}
}

func NewRegistry(srv, secret string, s *Server) *Registry {
	return &Registry{
		srv:        srv,
		secret:     secret,
		server:     s,
		hbchan:     make(chan struct{}),
		reportchan: make(chan struct{}),
	}
}

func (r *Registry) Run() error {
	go r.heartbeat()
	go r.report()
	for {
		r.run()
		time.Sleep(time.Second * 3)
	}
}

func (r *Registry) run() error {
	conn, err := net.DialTimeout("tcp", r.srv, time.Second*30)
	if err != nil {
		log.Error("%v", err)
		return err
	}

	defer conn.Close()

	reg := codec.RegisterReq{
		SecretKey: r.secret,
		PublicIP:  os.Getenv("PUBLIC_IP"),
	}
	err = codec.WriteJSON(conn, codec.CmdRegister, &reg)
	if err != nil {
		log.Error("write json: %v", err)
		return err
	}

	reply := &codec.RegisterReply{}
	codec.ReadJSON(conn, reply)
	log.Debug("%v", reply)
	if reply.CSPInfo == nil {
		log.Error("get csp fail: %v", reply)
		return fmt.Errorf("get csp fail")
	}

	instance, err := vpc.GetVPCInstance(reply.CSPInfo.CspType, reply.CSPInfo.AccessKey, reply.CSPInfo.AccessSecret)
	if err != nil {
		log.Error("unsupported vpc %v", reply.CSPInfo.CspType)
		// return err
	} else {
		r.server.SetVPCInstance(instance)
	}
	r.server.AddPeers(reply.EdgeList)

	go r.read(conn)
	r.write(conn)
	return nil
}

func (r *Registry) report() {
	tick := time.NewTicker(time.Second * 30)
	defer tick.Stop()
	for range tick.C {
		select {
		case r.reportchan <- struct{}{}:
		default:
		}
	}
}

func (r *Registry) heartbeat() {
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for range tick.C {
		select {
		case r.hbchan <- struct{}{}:
		default:
		}
	}
}

func (r *Registry) write(conn net.Conn) {
	for {
		select {
		case <-r.hbchan:
			log.Debug("send heartbeat to server")
			hb := &codec.Heartbeat{}
			conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
			err := codec.WriteJSON(conn, codec.CmdHeartbeat, hb)
			conn.SetWriteDeadline(time.Time{})
			if err != nil {
				log.Error("invalid hb msg: %v", err)
				return
			}
		case <-r.reportchan:
			report := ResetStat()
			conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
			err := codec.WriteJSON(conn, codec.CmdReport, report)
			if err != nil {
				log.Error("write json fail: %v", err)
			}
			conn.SetWriteDeadline(time.Time{})
		}
	}
}

func (r *Registry) read(conn net.Conn) {
	for {
		hdr, body, err := codec.Read(conn)
		if err != nil {
			log.Error("read fail: %v", err)
			return
		}

		switch hdr.Cmd() {
		case codec.CmdHeartbeat:
			log.Debug("heartbeat from server ")

		case codec.CmdAdd:
			log.Debug("online cmd: %s", string(body))
			online := codec.BroadcastOnlineMsg{}
			err := json.Unmarshal(body, &online)
			if err != nil {
				log.Error("invalid online msg %v", err)
				continue
			}
			r.server.AddPeer(&codec.Edge{
				ListenAddr: online.ListenAddr,
				Cidr:       online.Cidr,
			})

		case codec.CmdDel:
			log.Info("offline cmd: %s", string(body))
			offline := codec.BroadcastOfflineMsg{}
			err := json.Unmarshal(body, &offline)
			if err != nil {
				log.Error("invalid offline msg %v ", err)
				continue
			}
			r.server.DelPeer(&codec.Edge{
				ListenAddr: offline.ListenAddr,
				Cidr:       offline.Cidr,
			})

		case codec.CmdAddRoute:
			log.Debug("add route cmd: %s", string(body))
			addRoute := codec.AddRouteMsg{}
			err := json.Unmarshal(body, &addRoute)
			if err != nil {
				log.Error("invalid add route msg: %v", err)
				continue
			}
			r.server.AddRoute(&addRoute)

		case codec.CmdDelRoute:
			log.Debug("del route cmd: %s", string(body))
			delRoute := codec.DelRouteMsg{}
			err := json.Unmarshal(body, &delRoute)
			if err != nil {
				log.Error("invalid add route msg: %v", err)
				continue
			}
			r.server.DelRoute(&delRoute)

		}
	}
}

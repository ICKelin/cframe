package main

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	log "github.com/ICKelin/cframe/pkg/logs"
)

type Registry struct {
	srv    string
	name   string
	secret string
	server *Server

	//heart beat channel
	hbchan chan struct{}

	// report channel
	reportchan chan struct{}

	// report channel store msg
	// to be reported.
	// will drop the overflow reported msg
	mu          sync.Mutex
	reportQueue []string
	reporting   map[string]struct{}
}

func NewRegistry(srv, name, secret string, s *Server) *Registry {
	return &Registry{
		srv:         srv,
		name:        name,
		secret:      secret,
		server:      s,
		hbchan:      make(chan struct{}),
		reportchan:  make(chan struct{}),
		reportQueue: make([]string, 0),
		reporting:   make(map[string]struct{}),
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
		Name:      r.name,
		SecretKey: r.secret,
	}
	err = codec.WriteJSON(conn, codec.CmdRegister, &reg)
	if err != nil {
		log.Error("write json: %v", err)
		return err
	}

	reply := &codec.RegisterReply{}
	codec.ReadJSON(conn, reply)
	log.Debug("%v", reply)
	r.server.AddPeers(reply.OnlineHost)

	go r.read(conn)
	r.write(conn)
	return nil
}

func (r *Registry) heartbeat() {
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for range tick.C {
		r.hbchan <- struct{}{}
	}
}

func (r *Registry) report() {
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for range tick.C {
		timeout := time.NewTicker(time.Second * 3)
		select {
		case r.reportchan <- struct{}{}:
		case <-timeout.C:
			log.Warn("report channel timeout")
		}
		timeout.Stop()
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
			r.mu.Lock()
			if len(r.reportQueue) <= 0 {
				r.mu.Unlock()
				continue
			}
			q := make([]string, len(r.reportQueue))
			copy(q, r.reportQueue)
			r.reportQueue = r.reportQueue[:0]
			r.reporting = make(map[string]struct{})
			r.mu.Unlock()

			report := codec.ReportEdgeHost{
				HostIPs: q,
			}

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
			r.server.AddPeer(&codec.Host{
				HostAddr: online.HostAddr,
				Cidr:     online.Cidr,
			})

		case codec.CmdDel:
			log.Info("offline cmd: %s", string(body))
			offline := codec.BroadcastOfflineMsg{}
			err := json.Unmarshal(body, &offline)
			if err != nil {
				log.Error("invalid offline msg %v ", err)
				continue
			}
			r.server.DelPeer(&codec.Host{
				HostAddr: offline.HostAddr,
				Cidr:     offline.Cidr,
			})
		}
	}
}

// add report to report channel
func (r *Registry) Report(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reporting[ip]; ok {
		return
	}

	ipv4 := net.ParseIP(ip)
	if ipv4.To4() == nil {
		log.Warn("ignore invalid ip %s", ip)
		return
	}

	if ipv4.IsLoopback() || ipv4.Equal(net.IPv4zero) {
		return
	}

	log.Info("add report ip %s", ip)
	r.reportQueue = append(r.reportQueue, ip)
	r.reporting[ip] = struct{}{}
}

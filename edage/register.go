package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/ICKelin/cframe/codec"
)

type Registry struct {
	srv      string
	hostAddr string
	cidr     string
	server   *Server
}

func NewRegistry(srv, host, cidr string, s *Server) *Registry {
	return &Registry{
		srv:      srv,
		hostAddr: host,
		cidr:     cidr,
		server:   s,
	}
}

func (r *Registry) Run() error {
	for {
		r.run()
		time.Sleep(time.Second * 3)
	}
}

func (r *Registry) run() error {
	conn, err := net.DialTimeout("tcp", r.srv, time.Second*30)
	if err != nil {
		log.Println("[E] ", err)
		return err
	}

	defer conn.Close()

	reg := codec.RegisterReq{
		HostAddr:      r.hostAddr,
		ContainerCidr: r.cidr,
	}
	err = codec.WriteJSON(conn, codec.CmdRegister, &reg)
	if err != nil {
		log.Println("[E] ", err)
		return err
	}

	reply := &codec.RegisterReply{}
	codec.ReadJSON(conn, reply)
	r.server.AddPeers(reply.OnlineHost)

	hbchan := make(chan struct{})
	go r.read(conn)
	go r.heartbeat(hbchan)
	r.write(conn, hbchan)
	return nil

}

func (r *Registry) heartbeat(hbchan chan struct{}) {
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for range tick.C {
		hbchan <- struct{}{}
	}
}

func (r *Registry) write(conn net.Conn, hbchan chan struct{}) {
	for {
		select {
		case <-hbchan:
			hb := &codec.Heartbeat{}
			err := codec.WriteJSON(conn, codec.CmdHeartbeat, hb)
			if err != nil {
				log.Println("[E] ", err)
				return
			}
		}
	}
}

func (r *Registry) read(conn net.Conn) {
	for {
		hdr, body, err := codec.Read(conn)
		if err != nil {
			log.Println(err)
			return
		}

		switch hdr.Cmd() {
		case codec.CmdHeartbeat:
			log.Println("[I] heartbeat from server ")

		case codec.CmdOnline:
			log.Println("online cmd: ", string(body))
			online := codec.BroadcastOnlineMsg{}
			err := json.Unmarshal(body, &online)
			if err != nil {
				log.Println("[E] ", err)
				continue
			}
			r.server.AddPeer(&codec.Host{
				HostAddr:      online.HostAddr,
				ContainerCidr: online.ContainerCidr,
			})

		case codec.CmdOffline:
			log.Println("offline cmd: ", string(body))
			offline := codec.BroadcastOfflineMsg{}
			err := json.Unmarshal(body, &offline)
			if err != nil {
				log.Println("[E] ", err)
				continue
			}
			r.server.DelPeer(&codec.Host{
				HostAddr:      offline.HostAddr,
				ContainerCidr: offline.ContainerCidr,
			})
		}
	}
}

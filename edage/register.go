package main

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
)

type Registry struct {
	srv    string
	name   string
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

func NewRegistry(srv, name string, s *Server) *Registry {
	return &Registry{
		srv:         srv,
		name:        name,
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
		log.Println("[E] ", err)
		return err
	}

	defer conn.Close()

	reg := codec.RegisterReq{
		Name: r.name,
	}
	err = codec.WriteJSON(conn, codec.CmdRegister, &reg)
	if err != nil {
		log.Println("[E] ", err)
		return err
	}

	reply := &codec.RegisterReply{}
	codec.ReadJSON(conn, reply)
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
	tick := time.NewTicker(time.Minute * 1)
	defer tick.Stop()

	for range tick.C {
		r.reportchan <- struct{}{}
	}
}

func (r *Registry) write(conn net.Conn) {
	for {
		select {
		case <-r.hbchan:
			hb := &codec.Heartbeat{}
			err := codec.WriteJSON(conn, codec.CmdHeartbeat, hb)
			if err != nil {
				log.Println("[E] ", err)
				return
			}
		case <-r.reportchan:
			r.mu.Lock()
			q := make([]string, len(r.reportQueue))
			copy(q, r.reportQueue)
			r.reportQueue = r.reportQueue[:0]
			r.reporting = make(map[string]struct{})
			r.mu.Unlock()

			report := codec.ReportEdageHost{
				HostIPs: q,
			}

			err := codec.WriteJSON(conn, codec.CmdReport, report)
			if err != nil {
				log.Println("[E] ", err)
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

		case codec.CmdAdd:
			log.Println("online cmd: ", string(body))
			online := codec.BroadcastOnlineMsg{}
			err := json.Unmarshal(body, &online)
			if err != nil {
				log.Println("[E] ", err)
				continue
			}
			r.server.AddPeer(&codec.Host{
				HostAddr: online.HostAddr,
				Cidr:     online.Cidr,
			})

		case codec.CmdDel:
			log.Println("offline cmd: ", string(body))
			offline := codec.BroadcastOfflineMsg{}
			err := json.Unmarshal(body, &offline)
			if err != nil {
				log.Println("[E] ", err)
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
		log.Println("[W] ignore invalid ip ", ip)
		return
	}

	log.Println("add report ip ", ip)
	r.reportQueue = append(r.reportQueue, ip)
	r.reporting[ip] = struct{}{}
}

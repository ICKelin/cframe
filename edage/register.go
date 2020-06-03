package main

import (
	"log"
	"net"
	"time"

	"github.com/ICKelin/cframe/codec"
)

type Registry struct {
	srv      string
	hostAddr string
	cidr     string
}

func NewRegistry(srv string, host, cidr string) *Registry {
	return &Registry{
		srv:      srv,
		hostAddr: host,
		cidr:     cidr,
	}
}

func (r *Registry) Run() error {
	return r.run()
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

	// TODO: read response

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
			log.Println("online cmd: ", body)

		case codec.CmdOffline:
			log.Println("offline cmd: ", body)

		}
	}
}

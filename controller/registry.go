package main

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/edagemanager"
)

// registry server for edages
// edages register information to registry server
// and keep connection alive
// once there is any edage online/offline
// registry will notify onlined edages the new edages info
type RegistryServer struct {
	// registry server listen tcp addr
	// eg: 0.0.0.0:58422
	addr string

	// online edages
	// key: edages host addr
	// val: edages info and tcp connection
	mu   sync.Mutex
	sess map[string]*Session
}

type Session struct {
	host *codec.Host
	conn net.Conn
}

func NewRegistryServer(addr string) *RegistryServer {
	return &RegistryServer{
		addr: addr,
		sess: make(map[string]*Session),
	}
}

func (s *RegistryServer) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer lis.Close()

	go s.state()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("[E] accept: ", err)
			return err
		}

		go s.onConn(conn)
	}
}

func (s *RegistryServer) onConn(conn net.Conn) {
	defer conn.Close()
	reg := codec.RegisterReq{}
	err := codec.ReadJSON(conn, &reg)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("[I] node register", reg)

	// verify edage
	// only if the edages is configured with api server
	// or controller build in configuration
	// then the edages is valid
	edage := edagemanager.GetEdage(reg.Name)
	if edage == nil {
		log.Printf("[E] get edage for %s fail\n", reg.Name)
		return
	}

	log.Printf("[I] register success: %v\n", edage)

	host := edage.HostAddr
	onlineHosts := make([]*codec.Host, 0)
	s.mu.Lock()
	for _, sess := range s.sess {
		if sess.host.HostAddr != host {
			onlineHosts = append(onlineHosts, sess.host)
		}
	}
	s.sess[host] = &Session{
		host: &codec.Host{
			HostAddr: host,
			Cidr:     edage.Cidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess, host)
		s.mu.Unlock()
	}()

	// response current online edages
	err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
		OnlineHost: onlineHosts,
	})
	if err != nil {
		log.Println("[E] ", err)
	}

	// broadcast edage online to all onlined edages
	s.broadCastOnline(edage)

	// broadcast edage offline to all onlined edages
	defer s.broadcastOffline(edage)

	// keepalived...
	fail := 0
	hb := codec.Heartbeat{}
	for {
		header, body, err := codec.Read(conn)
		if err != nil {
			log.Println(err)
			fail += 1
			if fail >= 3 {
				break
			}
			time.Sleep(time.Second * 1)
			continue
		}

		switch header.Cmd() {
		case codec.CmdHeartbeat:
			// log.Println("heartbeat from client: ", conn.RemoteAddr().String())
			err = codec.WriteJSON(conn, codec.CmdHeartbeat, &hb)
			if err != nil {
				log.Println(err)
			}

		case codec.CmdReport:
			log.Println("receive report msg from edage: ", edage.Name)
			reportMsg := codec.ReportEdageHost{}
			err = json.Unmarshal(body, &reportMsg)
			if err != nil {
				log.Println("[E] ", err)
				continue
			}

			for _, ip := range reportMsg.HostIPs {
				host := &edagemanager.EdageHost{
					IP: ip,
				}

				edagemanager.AddedageHost(edage, host)
			}
		default:
			log.Println("unsupported cmd ", header.Cmd())
		}

		fail = 0
	}
}

func (s *RegistryServer) broadCastOnline(edage *edagemanager.Edage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess {
		if addr == edage.HostAddr {
			continue
		}

		go s.online(host.conn, edage)
	}
}

func (s *RegistryServer) online(peer net.Conn, edage *edagemanager.Edage) {
	log.Printf("[I] send online msg %v to %s\b",
		edage, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		HostAddr: edage.HostAddr,
		Cidr:     edage.Cidr,
	}

	err := codec.WriteJSON(peer, codec.CmdAdd, obj)
	if err != nil {
		log.Println("[E] ", err)
	}
}

func (s *RegistryServer) broadcastOffline(edage *edagemanager.Edage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess {
		if addr == edage.HostAddr {
			continue
		}

		go s.offline(host.conn, edage)
	}
}

func (s *RegistryServer) offline(peer net.Conn, edage *edagemanager.Edage) {
	log.Printf("[I] send offline msg %v to %s\b",
		edage, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		HostAddr: edage.HostAddr,
		Cidr:     edage.Cidr,
	}

	err := codec.WriteJSON(peer, codec.CmdDel, obj)
	if err != nil {
		log.Println("[E] ", err)
	}
}

func (s *RegistryServer) state() {
	tick := time.NewTicker(time.Second * 30)
	defer tick.Stop()
	for range tick.C {
		s.mu.Lock()
		for _, sess := range s.sess {
			log.Printf("%v\n", sess.host)
		}
		s.mu.Unlock()
	}
}

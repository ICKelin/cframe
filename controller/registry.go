package main

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/edagemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
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
			log.Error("accept: ", err)
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
		log.Error("read json fail: %v", err)
		return
	}

	log.Info("node register %v", reg)

	// verify edage
	// only if the edages is configured with api server
	// or controller build in configuration
	// then the edages is valid
	edage := edagemanager.GetEdage(reg.Name)
	if edage == nil {
		log.Error("get edage for %s fail\n", reg.Name)
		return
	}

	log.Info("register success: %v", edage)

	host := edage.HostAddr

	onlineHosts := make([]*codec.Host, 0)
	edages := edagemanager.GetEdages()
	for _, edg := range edages {
		if edg.HostAddr != host {
			onlineHosts = append(onlineHosts, &codec.Host{
				HostAddr: edg.HostAddr,
				Cidr:     edg.Cidr,
			})
		}
	}

	s.mu.Lock()
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
		log.Error("write json fail: %v", err)
		return
	}

	// keepalived...
	fail := 0
	hb := codec.Heartbeat{}
	for {
		header, body, err := codec.Read(conn)
		if err != nil {
			log.Error("read fail: %v", err)
			fail += 1
			if fail >= 3 {
				break
			}
			time.Sleep(time.Second * 1)
			continue
		}

		switch header.Cmd() {
		case codec.CmdHeartbeat:
			log.Debug("heartbeat from client: %s", conn.RemoteAddr().String())
			err = codec.WriteJSON(conn, codec.CmdHeartbeat, &hb)
			if err != nil {
				log.Error("write json fail: %v", err)
			}

		case codec.CmdReport:
			log.Info("receive report msg from edage: %s", edage.Name)
			reportMsg := codec.ReportEdageHost{}
			err = json.Unmarshal(body, &reportMsg)
			if err != nil {
				log.Error("invalid report msg: %v", err)
				continue
			}

			for _, ip := range reportMsg.HostIPs {
				host := &edagemanager.EdageHost{
					IP: ip,
				}

				edagemanager.AddedageHost(edage, host)
			}
		default:
			log.Warn("unsupported cmd %d", header.Cmd())
		}

		fail = 0
	}
}

func (s *RegistryServer) broadcastOnline(edage *edagemanager.Edage) {
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
	log.Info("[I] send online msg %v to %s",
		edage, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		HostAddr: edage.HostAddr,
		Cidr:     edage.Cidr,
	}

	err := codec.WriteJSON(peer, codec.CmdAdd, obj)
	if err != nil {
		log.Error("write json fail: %v", err)
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
	log.Info("send offline msg %v to %s\n",
		edage, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		HostAddr: edage.HostAddr,
		Cidr:     edage.Cidr,
	}

	err := codec.WriteJSON(peer, codec.CmdDel, obj)
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) state() {
	tick := time.NewTicker(time.Second * 30)
	defer tick.Stop()
	for range tick.C {
		s.mu.Lock()
		for _, sess := range s.sess {
			log.Info("edage: %s cidr: %s", sess.host.HostAddr, sess.host.Cidr)
		}
		s.mu.Unlock()
	}
}

func (s *RegistryServer) DelEdage(edg *edagemanager.Edage) {
	log.Info("delete edage: %v", edg)
	s.broadcastOffline(edg)
	// force edage connection offline
	edgSess := s.sess[edg.HostAddr]
	if edgSess != nil {
		log.Info("force close edage connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()

	}
}

func (s *RegistryServer) ModifyEdage(edg *edagemanager.Edage) {
	log.Info("modify edage: %v", edg)
	s.broadcastOnline(edg)
}

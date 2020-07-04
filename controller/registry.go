package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/edagemanager"
)

type RegistryServer struct {
	addr string

	// 所有host信息
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

	// 响应
	err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
		OnlineHost: onlineHosts,
	})
	if err != nil {
		log.Println("[E] ", err)
	}

	// 通知上线
	s.broadCastOnline(edage)

	// 通知下线
	defer s.broadcastOffline(edage)

	// 维持心跳
	fail := 0
	for {
		hb := codec.Heartbeat{}
		err := codec.ReadJSON(conn, &hb)
		if err != nil {
			log.Println(err)
			fail += 1
			if fail >= 3 {
				break
			}
			time.Sleep(time.Second * 1)
			continue
		}

		// log.Println("heartbeat from client: ", conn.RemoteAddr().String())
		err = codec.WriteJSON(conn, codec.CmdHeartbeat, &hb)
		if err != nil {
			log.Println(err)
		}

		fail = 0
	}
}

// 广播节点上线
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

// 广播节点下线
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

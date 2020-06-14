package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
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

	onlineHosts := make([]*codec.Host, 0)
	s.mu.Lock()
	for _, sess := range s.sess {
		if sess.host.HostAddr != reg.HostAddr {
			onlineHosts = append(onlineHosts, sess.host)
		}
	}

	s.sess[reg.HostAddr] = &Session{
		host: &codec.Host{
			HostAddr:      reg.HostAddr,
			ContainerCidr: reg.ContainerCidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess, reg.HostAddr)
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
	s.broadCastOnline(&reg)

	// 通知下线
	defer s.broadcastOffline(&reg)

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
func (s *RegistryServer) broadCastOnline(reg *codec.RegisterReq) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess {
		if addr == reg.HostAddr {
			continue
		}

		go s.online(host.conn, reg)
	}
}

func (s *RegistryServer) online(peer net.Conn, host *codec.RegisterReq) {
	log.Printf("[I] send online msg %v to %s\b",
		host, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		HostAddr:      host.HostAddr,
		ContainerCidr: host.ContainerCidr,
	}

	err := codec.WriteJSON(peer, codec.CmdOnline, obj)
	if err != nil {
		log.Println("[E] ", err)
	}
}

// 广播节点下线
func (s *RegistryServer) broadcastOffline(reg *codec.RegisterReq) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess {
		if addr == reg.HostAddr {
			continue
		}

		go s.offline(host.conn, reg)
	}
}

func (s *RegistryServer) offline(peer net.Conn, host *codec.RegisterReq) {
	log.Printf("[I] send offline msg %v to %s\b",
		host, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		HostAddr:      host.HostAddr,
		ContainerCidr: host.ContainerCidr,
	}

	err := codec.WriteJSON(peer, codec.CmdOffline, obj)
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

package main

import (
	"log"
	"net"
	"sync"

	"github.com/ICKelin/cframe/codec"
)

type RegistryServer struct {
	addr string

	// 所有host信息
	mu    sync.Mutex
	hosts map[string]*Host
}

type Host struct {
	addr string
	cidr string
	conn net.Conn
}

func NewRegistryServer(addr string) *RegistryServer {
	return &RegistryServer{
		addr:  addr,
		hosts: make(map[string]*Host),
	}
}

func (s *RegistryServer) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer lis.Close()

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

	s.mu.Lock()
	s.hosts[reg.HostAddr] = &Host{
		addr: reg.HostAddr,
		cidr: reg.ContainerCidr,
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.hosts, reg.HostAddr)
		s.mu.Unlock()
	}()

	log.Println("[I] node register", reg)

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

		log.Println("heartbeat from client: ", conn.RemoteAddr().String())
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
	for addr, host := range s.hosts {
		if addr == reg.HostAddr {
			continue
		}

		go s.online(host.conn, reg)
	}
}

func (s *RegistryServer) online(peer net.Conn, host *codec.RegisterReq) {

}

// 广播节点下线
func (s *RegistryServer) broadcastOffline(reg *codec.RegisterReq) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.hosts {
		if addr == reg.HostAddr {
			continue
		}

		go s.offline(host.conn, reg)
	}
}

func (s *RegistryServer) offline(conn net.Conn, reg *codec.RegisterReq) {

}

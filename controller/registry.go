package main

import (
	"log"
	"net"
	"sync"

	"github.com/ICKelin/cframe/inner_proto"
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

func NewRegistryServer() *RegistryServer {
	return &RegistryServer{}
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
	reg := inner_proto.RegisterReq{}
	err := ReadJSON(conn, &reg)
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

	// 通知上线
	s.broadCastOnline(&reg)

	// 通知下线
	defer s.broadcastOffline(&reg)

	// 维持心跳
	fail := 0
	for {
		hb := inner_proto.Heartbeat{}
		err := ReadJSON(conn, &hb)
		if err != nil {
			log.Println(err)
			fail += 1
			if fail >= 3 {
				break
			}
			continue
		}

		log.Println("heartbeat from client: ", conn.RemoteAddr().String())
		err = WriteJSON(conn, CmdHeartbeat, &hb)
		if err != nil {
			log.Println(err)
		}

		fail = 0
	}
}

// 广播节点上线
func (s *RegistryServer) broadCastOnline(reg *inner_proto.RegisterReq) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.hosts {
		if addr == reg.HostAddr {
			continue
		}

		go s.online(host.conn, reg)
	}
}

func (s *RegistryServer) online(peer net.Conn, host *inner_proto.RegisterReq) {

}

// 广播节点下线
func (s *RegistryServer) broadcastOffline(reg *inner_proto.RegisterReq) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.hosts {
		if addr == reg.HostAddr {
			continue
		}

		go s.offline(host.conn, reg)
	}
}

func (s *RegistryServer) offline(conn net.Conn, reg *inner_proto.RegisterReq) {

}

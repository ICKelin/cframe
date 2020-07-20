package main

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/pkg/edgemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
)

// registry server for edges
// edges register information to registry server
// and keep connection alive
// once there is any edge online/offline
// registry will notify onlined edges the new edges info
type RegistryServer struct {
	// registry server listen tcp addr
	// eg: 0.0.0.0:58422
	addr string

	// online edges
	// key: edges host addr
	// val: edges info and tcp connection
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

	// verify edge
	// only if the edges is configured with api server
	// or controller build in configuration
	// then the edges is valid
	edge := edgemanager.GetEdge(reg.Name)
	if edge == nil {
		log.Error("get edge for %s fail\n", reg.Name)
		return
	}

	log.Info("register success: %v", edge)

	host := edge.HostAddr

	onlineHosts := make([]*codec.Host, 0)
	edges := edgemanager.GetEdges()
	for _, edg := range edges {
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
			Cidr:     edge.Cidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess, host)
		s.mu.Unlock()
	}()

	// response current online edges
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
			log.Info("receive report msg from edge: %s", edge.Name)
			reportMsg := codec.ReportEdgeHost{}
			err = json.Unmarshal(body, &reportMsg)
			if err != nil {
				log.Error("invalid report msg: %v", err)
				continue
			}

			for _, ip := range reportMsg.HostIPs {
				host := &edgemanager.EdgeHost{
					IP: ip,
				}

				edgemanager.AddedgeHost(edge, host)
			}
		default:
			log.Warn("unsupported cmd %d", header.Cmd())
		}

		fail = 0
	}
}

func (s *RegistryServer) broadcastOnline(edge *edgemanager.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess {
		if addr == edge.HostAddr {
			continue
		}

		go s.online(host.conn, edge)
	}
}

func (s *RegistryServer) online(peer net.Conn, edge *edgemanager.Edge) {
	log.Info("[I] send online msg %v to %s",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		HostAddr: edge.HostAddr,
		Cidr:     edge.Cidr,
	}

	err := codec.WriteJSON(peer, codec.CmdAdd, obj)
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) broadcastOffline(edge *edgemanager.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess {
		if addr == edge.HostAddr {
			continue
		}

		go s.offline(host.conn, edge)
	}
}

func (s *RegistryServer) offline(peer net.Conn, edge *edgemanager.Edge) {
	log.Info("send offline msg %v to %s\n",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		HostAddr: edge.HostAddr,
		Cidr:     edge.Cidr,
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
			log.Info("edge: %s cidr: %s", sess.host.HostAddr, sess.host.Cidr)
		}
		s.mu.Unlock()
	}
}

func (s *RegistryServer) DelEdge(edg *edgemanager.Edge) {
	log.Info("delete edge: %v", edg)
	s.broadcastOffline(edg)
	// force edge connection offline
	edgSess := s.sess[edg.HostAddr]
	if edgSess != nil {
		log.Info("force close edge connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()

	}
}

func (s *RegistryServer) ModifyEdge(edg *edgemanager.Edge) {
	log.Info("modify edge: %v", edg)
	s.broadcastOnline(edg)
}

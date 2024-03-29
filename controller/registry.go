package main

import (
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/models"
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
	// key: edge listen addr
	// val: edge info and tcp connection
	mu   sync.Mutex
	sess map[string]map[string]*Session

	// edge manager
	edgeManager *models.EdgeManager

	// route manager
	routeManager *models.RouteManager

	// namespace manager
	namespaceMgr *models.NamespaceManager
}

type Session struct {
	edge *codec.Edge
	conn net.Conn
}

func NewRegistryServer(addr string,
	edgeMgr *models.EdgeManager,
	routeMgr *models.RouteManager,
	namespaceMgr *models.NamespaceManager) *RegistryServer {
	return &RegistryServer{
		addr:         addr,
		sess:         make(map[string]map[string]*Session),
		edgeManager:  edgeMgr,
		routeManager: routeMgr,
		namespaceMgr: namespaceMgr,
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

	log.Info("edge register %+v", reg)

	// verify namespace
	nsInfo, err := s.namespaceMgr.GetNamespace(reg.Namespace)
	if err != nil {
		log.Error("get namespace %s fail: %v", reg.Namespace, err)
		return
	}

	if nsInfo.Secret != reg.SecretKey {
		log.Error("verify namespace key fail")
		return
	}

	log.Info("namespace info: %+v", nsInfo)

	// verify edge
	edges := s.edgeManager.GetEdges(nsInfo.Name)
	if len(edges) <= 0 {
		log.Error("get edges for namespace %s fail", nsInfo.Name)
		return
	}

	otherEdges := make([]*codec.Edge, 0, len(edges)-1)
	curEdge := &codec.Edge{}

	find := false
	for i, edge := range edges {
		if edge.Name == reg.Name {
			find = true
			curEdge = edges[i]
			continue
		}

		otherEdges = append(otherEdges, edges[i])
	}
	if !find {
		log.Error("verify edge fail, edge not in %s namespace", nsInfo.Name)
		return
	}

	log.Info("other edge list: %+v", otherEdges)

	// TODO: get csp info

	// get routes
	routes := s.routeManager.GetRoutes(nsInfo.Name)
	log.Info("route list: %+v", routes)
	otherRoutes := make([]*codec.Route, 0)
	for i, route := range routes {
		if route.Nexthop == curEdge.ListenAddr {
			continue
		}
		otherRoutes = append(otherRoutes, routes[i])
	}
	log.Info("will dispatch route list: ", otherRoutes)

	// store session
	sessKey := nsInfo.Name
	s.mu.Lock()
	if s.sess[sessKey] == nil {
		s.sess[sessKey] = make(map[string]*Session)
	}
	if _, ok := s.sess[sessKey][curEdge.ListenAddr]; ok {
		log.Warn("edge %s addr %s is running", curEdge.Name, curEdge.ListenAddr)
		s.mu.Unlock()
		return
	}

	s.sess[sessKey][curEdge.ListenAddr] = &Session{
		edge: &codec.Edge{
			ListenAddr: curEdge.ListenAddr,
			Cidr:       curEdge.Cidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess[sessKey], curEdge.ListenAddr)
		s.mu.Unlock()
	}()

	// reply to edge
	conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
		EdgeList: otherEdges,
		Routes:   otherRoutes,
	})
	conn.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write json fail: %v", err)
		return
	}

	// keepalived
	fail := 0
	hb := codec.Heartbeat{}
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * 30))
		header, body, err := codec.Read(conn)
		conn.SetReadDeadline(time.Time{})
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
			log.Debug("receive report from edge: %s %s", curEdge.Name, string(body))

		case codec.CmdAlarm:
			log.Info("receive alarm from edge: %s %s", curEdge.Name, string(body))

		default:
			log.Warn("unsupported cmd %d", header.Cmd())
		}

		fail = 0
	}
}

func (s *RegistryServer) broadcastOnline(namespace string, edge *codec.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess[namespace] {
		if addr == edge.ListenAddr {
			continue
		}

		go s.online(host.conn, edge)
	}
}

func (s *RegistryServer) online(peer net.Conn, edge *codec.Edge) {
	log.Info("[I] send online msg %v to %s",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		ListenAddr: edge.ListenAddr,
		Cidr:       edge.Cidr,
	}

	peer.SetWriteDeadline(time.Now().Add(time.Second * 10))
	err := codec.WriteJSON(peer, codec.CmdAdd, obj)
	peer.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) broadcastOffline(namespace string, edge *codec.Edge) {
	s.mu.Lock()
	var conn net.Conn
	for addr, host := range s.sess[namespace] {
		if addr == edge.ListenAddr {
			conn = host.conn
			continue
		}

		go s.offline(host.conn, edge)
	}
	s.mu.Unlock()

	// exit to stop edge process
	if conn != nil {
		conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		codec.WriteJSON(conn, codec.CmdExit, nil)
		conn.SetWriteDeadline(time.Time{})
	}
}

func (s *RegistryServer) offline(peer net.Conn, edge *codec.Edge) {
	log.Info("send offline msg %v to %s\n",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		ListenAddr: edge.ListenAddr,
		Cidr:       edge.Cidr,
	}

	peer.SetWriteDeadline(time.Now().Add(time.Second * 10))
	err := codec.WriteJSON(peer, codec.CmdDel, obj)
	peer.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) broadcastAddRoute(namespace string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[namespace] {
		if addr == r.Nexthop {
			continue
		}

		go s.addRoute(host.conn, r)
	}
}

func (s *RegistryServer) addRoute(peer net.Conn, r *codec.Route) {
	log.Info("send addroute msg %v to %s\n",
		r, peer.RemoteAddr().String())

	obj := &codec.AddRouteMsg{
		Cidr:    r.CIDR,
		Nexthop: r.Nexthop,
	}

	peer.SetWriteDeadline(time.Now().Add(time.Second * 10))
	err := codec.WriteJSON(peer, codec.CmdAddRoute, obj)
	peer.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) broadcastDelRoute(namespace string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[namespace] {
		if addr == r.Nexthop {
			continue
		}

		go s.delRoute(host.conn, r)
	}
}

func (s *RegistryServer) delRoute(peer net.Conn, r *codec.Route) {
	log.Info("send delroute msg %v to %s\n",
		r, peer.RemoteAddr().String())

	obj := &codec.DelRouteMsg{
		Cidr:    r.CIDR,
		Nexthop: r.Nexthop,
	}

	peer.SetWriteDeadline(time.Now().Add(time.Second * 10))
	err := codec.WriteJSON(peer, codec.CmdDelRoute, obj)
	peer.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write json fail: %v", err)
	}
}

func (s *RegistryServer) state() {
	tick := time.NewTicker(time.Second * 30)
	defer tick.Stop()
	for range tick.C {
		s.mu.Lock()
		for userId, sesses := range s.sess {
			for _, sess := range sesses {
				log.Info("namespace %s edge: %s cidr: %s",
					userId, sess.edge.ListenAddr, sess.edge.Cidr)
			}
		}
		s.mu.Unlock()
	}
}

func (s *RegistryServer) DelEdge(namespace string, edg *codec.Edge) {
	log.Info("delete edge: %s %v", namespace, edg)
	s.broadcastOffline(namespace, edg)
	// force edge connection offline
	edgSess := s.sess[namespace][edg.ListenAddr]
	if edgSess != nil {
		log.Info("force close edge connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()
	}
}

func (s *RegistryServer) ModifyEdge(namespace string, edg *codec.Edge) {
	log.Info("modify edge: %s %v", namespace, edg)
	s.broadcastOnline(namespace, edg)
}

func (s *RegistryServer) DelRoute(namespace string, route *codec.Route) {
	log.Info("del route: %s %v", namespace, route)
	s.broadcastDelRoute(namespace, route)
}

func (s *RegistryServer) AddRoute(namespace string, route *codec.Route) {
	log.Info("add route: %s %v", namespace, route)
	s.broadcastAddRoute(namespace, route)
}

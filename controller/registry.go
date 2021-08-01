package main

import (
	"net"
	"strings"
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

	// app manager
	appManager *models.AppManager
}

type Session struct {
	edge *codec.Edge
	conn net.Conn
}

func NewRegistryServer(addr string,
	edgeMgr *models.EdgeManager,
	routeMgr *models.RouteManager,
	appMgr *models.AppManager) *RegistryServer {
	return &RegistryServer{
		addr:         addr,
		sess:         make(map[string]map[string]*Session),
		edgeManager:  edgeMgr,
		routeManager: routeMgr,
		appManager:   appMgr,
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
	remoteIP := reg.PublicIP
	if len(remoteIP) <= 0 {
		remoteAddr := conn.RemoteAddr().String()
		remoteIP, _, _ = net.SplitHostPort(remoteAddr)
	}

	// verify key
	appInfo, err := s.appManager.GetApp(reg.SecretKey)
	if err != nil {
		log.Error("verify key fail: %v", err)
		return
	}
	log.Info("app info: %+v", appInfo)

	// verify edge
	edges := s.edgeManager.GetEdges(appInfo.Secret)
	otherEdges := make([]*codec.Edge, 0, len(edges)-1)
	curEdge := &codec.Edge{}

	find := false
	for i, edge := range edges {
		ipport := strings.Split(edge.ListenAddr, ":")
		if ipport[0] == remoteIP {
			find = true
			curEdge = edges[i]
			continue
		}

		otherEdges = append(otherEdges, edges[i])
	}
	if !find {
		log.Error("verify edge fail: %v", err)
		return
	}

	log.Info("other edge list: %+v", otherEdges)

	// TODO: get csp info

	// get routes
	routes := s.routeManager.GetRoutes(appInfo.Secret)
	log.Info("route list: %+v", routes)

	// reply to edge
	err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
		EdgeList: otherEdges,
		Routes:   routes,
	})
	if err != nil {
		log.Error("write json fail: %v", err)
		return
	}

	// store session
	s.mu.Lock()
	if s.sess[appInfo.Secret] == nil {
		s.sess[appInfo.Secret] = make(map[string]*Session)
	}
	s.sess[appInfo.Secret][curEdge.ListenAddr] = &Session{
		edge: &codec.Edge{
			ListenAddr: curEdge.ListenAddr,
			Cidr:       curEdge.Cidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess[appInfo.Secret], curEdge.ListenAddr)
		s.mu.Unlock()
	}()

	// keepalived
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
			log.Info("receive report from edge: %s %s", curEdge.Name, string(body))

		case codec.CmdAlarm:
			log.Info("receive alarm from edge: %s %s", curEdge.Name, string(body))

		default:
			log.Warn("unsupported cmd %d", header.Cmd())
		}

		fail = 0
	}
}

func (s *RegistryServer) broadcastOnline(appId string, edge *codec.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess[appId] {
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

func (s *RegistryServer) broadcastOffline(appId string, edge *codec.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[appId] {
		if addr == edge.ListenAddr {
			continue
		}

		go s.offline(host.conn, edge)
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

func (s *RegistryServer) broadcastAddRoute(appId string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[appId] {
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

func (s *RegistryServer) broadcastDelRoute(appId string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[appId] {
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
				log.Info("userId %s edge: %s cidr: %s",
					userId, sess.edge.ListenAddr, sess.edge.Cidr)
			}
		}
		s.mu.Unlock()
	}
}

func (s *RegistryServer) DelEdge(appId string, edg *codec.Edge) {
	log.Info("delete edge: %s %v", appId, edg)
	s.broadcastOffline(appId, edg)
	// force edge connection offline
	edgSess := s.sess[appId][edg.ListenAddr]
	if edgSess != nil {
		log.Info("force close edge connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()
	}
}

func (s *RegistryServer) ModifyEdge(appId string, edg *codec.Edge) {
	log.Info("modify edge: %s %v", appId, edg)
	s.broadcastOnline(appId, edg)
}

func (s *RegistryServer) DelRoute(appId string, route *codec.Route) {
	log.Info("del route: %s %v", appId, route)
	s.broadcastDelRoute(appId, route)
}

func (s *RegistryServer) AddRoute(appId string, route *codec.Route) {
	log.Info("add route: %s %v", appId, route)
	s.broadcastAddRoute(appId, route)
}

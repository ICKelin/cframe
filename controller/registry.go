package main

import (
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
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
}

type Session struct {
	edge *codec.Edge
	conn net.Conn
}

func NewRegistryServer(addr string) *RegistryServer {
	return &RegistryServer{
		addr: addr,
		sess: make(map[string]map[string]*Session),
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

	// TODO: verify edge

	// TODO: get csp info

	// TODO: get routes

	// TODO: reply to edge

	// // verify secret key
	// req := &proto.GetUserBySecretReq{
	// 	Secret: reg.SecretKey,
	// }
	// reply, err := s.userCli.GetUserBySecret(context.Background(), req)
	// if err != nil {
	// 	log.Error("get user by secret %sfail: %v", reg.SecretKey, err)
	// 	return
	// }
	// if reply.Code != 0 {
	// 	log.Error("get user by secret fail <%d>%s", reply.Code, reply.Message)
	// 	return
	// }

	// user := reply.UserInfo
	// if !bson.IsObjectIdHex(user.UserId) {
	// 	log.Error("invalid userId %s", user.UserId)
	// 	return
	// }
	// userObjectId := bson.ObjectIdHex(user.UserId)

	// log.Info("user info: %v", user)

	// // verify edge node
	// edgelist, err := s.edgeManager.GetEdgeList(userObjectId)
	// if err != nil {
	// 	log.Error("get edge list fail: %v", err)
	// 	return
	// }

	// find := false
	// var curEdge *models.EdgeInfo
	// otherEdge := make([]*models.EdgeInfo, 0)
	// for _, edge := range edgelist {
	// 	if edge.PublicIP == remoteIP {
	// 		find = true
	// 		curEdge = edge
	// 		continue
	// 	}

	// 	otherEdge = append(otherEdge, edge)
	// }
	// if !find {
	// 	log.Error("verify edge for ip %s fail, %s", remoteIP, user.UserId)
	// 	return
	// }

	// // get csp info
	// csp, err := s.cspManager.GetCSP(userObjectId, curEdge.CSPType)
	// if err != nil {
	// 	log.Error("get csp fail: %v", err)
	// 	// as build in csp
	// 	csp = &models.CSP{}
	// 	// return
	// }

	// // get routes info
	// routes := make([]*proto.Route, 0)
	// storeRoutes, _ := s.routeManager.GetOtherRoutes(userObjectId, fmt.Sprintf("%s:%d", curEdge.PublicIP, curEdge.PublicPort))
	// for _, r := range storeRoutes {
	// 	routes = append(routes, &proto.Route{
	// 		Cidr:    r.Cidr,
	// 		Nexthop: r.Nexthop,
	// 	})
	// }

	// log.Debug("csp info: %v", csp)
	// log.Info("register success: %v", curEdge)

	// edges := make([]*codec.Edge, 0)
	// for _, edg := range otherEdge {
	// 	edges = append(edges, &codec.Edge{
	// 		ListenAddr: fmt.Sprintf("%s:%d", edg.PublicIP, edg.PublicPort),
	// 		Cidr:       edg.Cidr,
	// 	})
	// }

	// s.mu.Lock()
	// curEdgeAddr := fmt.Sprintf("%s:%d", curEdge.PublicIP, curEdge.PublicPort)
	// if s.sess[userObjectId.Hex()] == nil {
	// 	s.sess[userObjectId.Hex()] = make(map[string]*Session)
	// }
	// s.sess[userObjectId.Hex()][curEdgeAddr] = &Session{
	// 	edge: &codec.Edge{
	// 		ListenAddr: curEdgeAddr,
	// 		Cidr:       curEdge.Cidr,
	// 	},
	// 	conn: conn,
	// }
	// s.mu.Unlock()
	// defer func() {
	// 	s.mu.Lock()
	// 	delete(s.sess[userObjectId.Hex()], curEdgeAddr)
	// 	s.mu.Unlock()
	// }()

	// // response current online edges
	// err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
	// 	EdgeList: edges,
	// 	CSPInfo: &codec.CSPInfo{
	// 		CspType:      csp.CSPType,
	// 		AccessKey:    csp.AccessKey,
	// 		AccessSecret: csp.SecretKey,
	// 	},
	// 	Routes: routes,
	// })
	// if err != nil {
	// 	log.Error("write json fail: %v", err)
	// 	return
	// }

	// // keepalived...
	// fail := 0
	// hb := codec.Heartbeat{}
	// for {
	// 	header, body, err := codec.Read(conn)
	// 	if err != nil {
	// 		log.Error("read fail: %v", err)
	// 		fail += 1
	// 		if fail >= 3 {
	// 			break
	// 		}
	// 		time.Sleep(time.Second * 1)
	// 		continue
	// 	}

	// 	switch header.Cmd() {
	// 	case codec.CmdHeartbeat:
	// 		log.Debug("heartbeat from client: %s", conn.RemoteAddr().String())
	// 		err = codec.WriteJSON(conn, codec.CmdHeartbeat, &hb)
	// 		if err != nil {
	// 			log.Error("write json fail: %v", err)
	// 		}
	// 		s.edgeManager.UpdateActive(userObjectId, curEdge.Name, time.Now())

	// 	case codec.CmdReport:
	// 		log.Info("receive report from edge: %s %s", curEdge.Name, string(body))
	// 		stat := codec.ReportMsg{}
	// 		json.Unmarshal(body, &stat)

	// 		s.statManager.AddStat(&models.Stat{
	// 			UserId:     userObjectId,
	// 			EdgeName:   curEdge.Name,
	// 			TrafficIn:  stat.TrafficIn,
	// 			TrafficOut: stat.TrafficOut,
	// 			Timestamp:  stat.Timestamp,
	// 		})

	// 	case codec.CmdAlarm:
	// 		log.Info("receive alarm from edge: %s %s", curEdge.Name, string(body))

	// 	default:
	// 		log.Warn("unsupported cmd %d", header.Cmd())
	// 	}

	// 	fail = 0
	// }
}

func (s *RegistryServer) broadcastOnline(userId string, edge *codec.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for addr, host := range s.sess[userId] {
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

func (s *RegistryServer) broadcastOffline(userId string, edge *codec.Edge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[userId] {
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

func (s *RegistryServer) broadcastAddRoute(userId string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[userId] {
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

func (s *RegistryServer) broadcastDelRoute(userId string, r *codec.Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, host := range s.sess[userId] {
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

func (s *RegistryServer) DelEdge(userId string, edg *codec.Edge) {
	log.Info("delete edge: %s %v", userId, edg)
	s.broadcastOffline(userId, edg)
	// force edge connection offline
	edgSess := s.sess[userId][edg.ListenAddr]
	if edgSess != nil {
		log.Info("force close edge connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()
	}
}

func (s *RegistryServer) ModifyEdge(userId string, edg *codec.Edge) {
	log.Info("modify edge: %s %v", userId, edg)
	s.broadcastOnline(userId, edg)
}

func (s *RegistryServer) DelRoute(userId string, route *codec.Route) {
	log.Info("del route: %s %v", userId, route)
	s.broadcastDelRoute(userId, route)
}

func (s *RegistryServer) AddRoute(userId string, route *codec.Route) {
	log.Info("add route: %s %v", userId, route)
	s.broadcastAddRoute(userId, route)
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/codec/proto"
	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/edgemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
	"gopkg.in/mgo.v2/bson"
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
	sess map[string]*Session

	// edge manager, query edge info from db
	edgeManager *models.EdgeManager

	// csp manager, query cloud service provicer from db
	cspManager *models.CSPManager

	// stat manager
	statManager *models.StatManager

	// user manager rpc client
	userCli proto.UserServiceClient
}

type Session struct {
	edge *codec.Edge
	conn net.Conn
}

func NewRegistryServer(addr string, cli proto.UserServiceClient) *RegistryServer {
	return &RegistryServer{
		addr:        addr,
		sess:        make(map[string]*Session),
		edgeManager: models.GetEdgeManager(),
		cspManager:  models.GetCSPManager(),
		statManager: models.GetStatManager(),
		userCli:     cli,
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
	remoteAddr := conn.RemoteAddr().String()
	remoteIP, _, _ := net.SplitHostPort(remoteAddr)

	// verify secret key
	req := &proto.GetUserBySecretReq{
		Secret: reg.SecretKey,
	}
	reply, err := s.userCli.GetUserBySecret(context.Background(), req)
	if err != nil {
		log.Error("get user by secret %sfail: %v", reg.SecretKey, err)
		return
	}
	if reply.Code != 0 {
		log.Error("get user by secret fail <%d>%s", reply.Code, reply.Message)
		return
	}

	user := reply.UserInfo
	if !bson.IsObjectIdHex(user.UserId) {
		log.Error("invalid userId %s", user.UserId)
		return
	}
	userObjectId := bson.ObjectIdHex(user.UserId)

	log.Info("user info: %v", user)

	// verify edge node
	edgelist, err := s.edgeManager.GetEdgeList(userObjectId)
	if err != nil {
		log.Error("get edge list fail: %v", err)
		return
	}

	find := false
	var curEdge *models.EdgeInfo
	otherEdge := make([]*models.EdgeInfo, 0)
	for _, edge := range edgelist {
		if edge.PublicIP == remoteIP {
			find = true
			curEdge = edge
			continue
		}

		otherEdge = append(otherEdge, edge)
	}
	if !find {
		log.Error("verify edge for ip %s fail, %s", remoteIP, user.UserId)
		return
	}

	// get csp info
	csp, err := s.cspManager.GetCSP(userObjectId, curEdge.CSPType)
	if err != nil {
		log.Error("get csp fail: %v", err)
		return
	}

	log.Debug("csp info: %v", csp)

	log.Info("register success: %v", curEdge)

	edges := make([]*codec.Edge, 0)
	for _, edg := range otherEdge {
		edges = append(edges, &codec.Edge{
			ListenAddr: fmt.Sprintf("%s:%d", edg.PublicIP, edg.PublicPort),
			Cidr:       edg.Cidr,
		})
	}

	s.mu.Lock()
	curEdgeAddr := fmt.Sprintf("%s:%d", curEdge.PublicIP, curEdge.PublicPort)
	s.sess[curEdgeAddr] = &Session{
		edge: &codec.Edge{
			ListenAddr: curEdgeAddr,
			Cidr:       curEdge.Cidr,
		},
		conn: conn,
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.sess, curEdgeAddr)
		s.mu.Unlock()
	}()

	// response current online edges
	err = codec.WriteJSON(conn, codec.CmdRegister, &codec.RegisterReply{
		EdgeList: edges,
		CSPInfo: &codec.CSPInfo{
			CspType:      csp.CSPType,
			AccessKey:    csp.AccessKey,
			AccessSecret: csp.SecretKey,
		},
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
			s.edgeManager.UpdateActive(userObjectId, curEdge.Name, time.Now())

		case codec.CmdReport:
			log.Info("receive report from edge: %s %s", curEdge.Comment, string(body))
			stat := codec.ReportMsg{}
			json.Unmarshal(body, &stat)

			s.statManager.AddStat(&models.Stat{
				UserId:     userObjectId,
				EdgeName:   curEdge.Name,
				TrafficIn:  stat.TrafficIn,
				TrafficOut: stat.TrafficOut,
				Timestamp:  stat.Timestamp,
			})

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
		if addr == edge.ListenAddr {
			continue
		}

		go s.online(host.conn, edge)
	}
}

func (s *RegistryServer) online(peer net.Conn, edge *edgemanager.Edge) {
	log.Info("[I] send online msg %v to %s",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOnlineMsg{
		ListenAddr: edge.ListenAddr,
		Cidr:       edge.Cidr,
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
		if addr == edge.ListenAddr {
			continue
		}

		go s.offline(host.conn, edge)
	}
}

func (s *RegistryServer) offline(peer net.Conn, edge *edgemanager.Edge) {
	log.Info("send offline msg %v to %s\n",
		edge, peer.RemoteAddr().String())

	obj := &codec.BroadcastOfflineMsg{
		ListenAddr: edge.ListenAddr,
		Cidr:       edge.Cidr,
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
			log.Info("edge: %s cidr: %s", sess.edge.ListenAddr, sess.edge.Cidr)
		}
		s.mu.Unlock()
	}
}

func (s *RegistryServer) DelEdge(edg *edgemanager.Edge) {
	log.Info("delete edge: %v", edg)
	s.broadcastOffline(edg)
	// force edge connection offline
	edgSess := s.sess[edg.ListenAddr]
	if edgSess != nil {
		log.Info("force close edge connection: %v", edgSess.conn.RemoteAddr())
		edgSess.conn.Close()

	}
}

func (s *RegistryServer) ModifyEdge(edg *edgemanager.Edge) {
	log.Info("modify edge: %v", edg)
	s.broadcastOnline(edg)
}

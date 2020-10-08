package main

import (
	"context"
	"net"

	"github.com/ICKelin/cframe/pkg/edgemanager"

	"github.com/ICKelin/cframe/codec/proto"
	"github.com/ICKelin/cframe/controller/models"
	log "github.com/ICKelin/cframe/pkg/logs"
	"google.golang.org/grpc"
	"gopkg.in/mgo.v2/bson"
)

type RPCServer struct {
	addr        string
	edgeManager *models.EdgeManager
	cspManager  *models.CSPManager
}

func NewRPCServer(addr string) *RPCServer {
	return &RPCServer{
		addr:        addr,
		edgeManager: models.GetEdgeManager(),
		cspManager:  models.GetCSPManager(),
	}
}

func (s *RPCServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Info("listenning %v success", s.addr)
	srv := grpc.NewServer()
	proto.RegisterControllerServiceServer(srv, s)
	return srv.Serve(listener)
}

func (s *RPCServer) GetEdgeList(ctx context.Context,
	req *proto.GetEdgeListReq) (*proto.GetEdgeListReply, error) {
	badReq := &proto.GetEdgeListReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	userId := bson.ObjectIdHex(req.UserId)
	edges, err := s.edgeManager.GetEdgeList(userId)
	if err != nil {
		return &proto.GetEdgeListReply{Code: 50000, Message: err.Error()}, nil
	}

	edgelist := make([]*proto.EdgeInfo, 0)
	for _, edge := range edges {
		item := &proto.EdgeInfo{
			CspType:    edge.CSPType,
			PublicIP:   edge.PublicIP,
			Cidr:       edge.Cidr,
			ListenAddr: edge.ListenAddr,
			Comment:    edge.Comment,
			UserId:     req.UserId,
			Name:       edge.Name,
			ActiveAt:   edge.ActiveAt,
		}
		edgelist = append(edgelist, item)
	}

	return &proto.GetEdgeListReply{Edges: edgelist}, nil
}

func (s *RPCServer) AddEdge(ctx context.Context,
	req *proto.AddEdgeReq) (*proto.AddEdgeReply, error) {
	badReq := &proto.AddEdgeReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	if len(req.Name) <= 0 {
		return &proto.AddEdgeReply{Code: 40000, Message: "invalid name"}, nil
	}

	log.Debug("add edge with req: %v", req)

	exist, _ := s.edgeManager.GetEdgeByName(bson.ObjectIdHex(req.UserId), req.Name)
	if exist != nil {
		return &proto.AddEdgeReply{Code: 50000, Message: "name exist"}, nil
	}

	edgeInfo := &models.EdgeInfo{
		Name:       req.Name,
		UserId:     bson.ObjectIdHex(req.UserId),
		CSPType:    req.CspType,
		PublicIP:   req.PublicIP,
		Cidr:       req.Cidr,
		ListenAddr: req.ListenAddr,
		Comment:    req.Comment,
	}

	r, err := s.edgeManager.AddEdge(edgeInfo)
	if err != nil {
		return &proto.AddEdgeReply{Code: 50000, Message: err.Error()}, nil
	}

	// add to etcd
	edgemanager.AddEdge(req.UserId, req.Name, &edgemanager.Edge{
		Name:       req.Name,
		Comment:    edgeInfo.Comment,
		Cidr:       edgeInfo.Cidr,
		ListenAddr: edgeInfo.ListenAddr,
		Type:       edgeInfo.CSPType,
	})

	return &proto.AddEdgeReply{
		Data: &proto.EdgeInfo{
			Name:       r.Name,
			CspType:    r.CSPType,
			UserId:     r.UserId.Hex(),
			PublicIP:   r.PublicIP,
			Cidr:       r.Cidr,
			ListenAddr: r.ListenAddr,
			Comment:    r.Comment,
		},
	}, nil
}

func (s *RPCServer) DelEdge(ctx context.Context,
	req *proto.DelEdgeReq) (*proto.DelEdgeReply, error) {
	badReq := &proto.DelEdgeReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	log.Debug("del edge with req: %v", req)
	// delete to etcd
	edgemanager.DelEdge(req.UserId, req.EdgeName)

	err := s.edgeManager.DelEdge(bson.ObjectIdHex(req.UserId), req.EdgeName)
	if err != nil {
		return &proto.DelEdgeReply{Code: 50000, Message: err.Error()}, nil
	}

	return &proto.DelEdgeReply{}, nil
}

func (s *RPCServer) GetCSPList(ctx context.Context,
	req *proto.GetCSPListReq) (*proto.GetCSPListReply, error) {
	badReq := &proto.GetCSPListReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	log.Debug("get csp with req: %v", req)
	userId := bson.ObjectIdHex(req.UserId)
	csps, err := s.cspManager.GetCSPList(userId)
	if err != nil {
		return &proto.GetCSPListReply{Code: 50000, Message: err.Error()}, nil
	}

	csplist := make([]*proto.CSPInfo, 0)
	for _, csp := range csps {
		item := &proto.CSPInfo{
			CspType:      csp.CSPType,
			AccessKey:    csp.AccessKey,
			AccessSecret: csp.SecretKey,
		}
		csplist = append(csplist, item)
	}

	return &proto.GetCSPListReply{CspInfo: csplist}, nil
}

func (s *RPCServer) AddCSP(ctx context.Context,
	req *proto.AddCSPReq) (*proto.AddCSPReply, error) {
	badReq := &proto.AddCSPReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	log.Debug("add csp with req: %v", req)
	if len(req.AccessKey) <= 0 ||
		len(req.AccessSecret) <= 0 {
		return &proto.AddCSPReply{Code: 50000, Message: "access key/secret must configured"}, nil
	}

	exist, _ := s.cspManager.GetCSP(bson.ObjectIdHex(req.UserId), req.CspType)
	if exist != nil {
		return &proto.AddCSPReply{Code: 50000, Message: "csp exist"}, nil
	}

	cspInfo := &models.CSP{
		UserId:    bson.ObjectIdHex(req.UserId),
		AccessKey: req.AccessKey,
		SecretKey: req.AccessSecret,
		CSPType:   req.CspType,
	}
	err := s.cspManager.AddCSP(cspInfo)
	if err != nil {
		return &proto.AddCSPReply{Code: 50000, Message: err.Error()}, nil
	}
	return &proto.AddCSPReply{Data: &proto.CSPInfo{
		CspType:      cspInfo.CSPType,
		AccessKey:    cspInfo.AccessKey,
		AccessSecret: cspInfo.SecretKey,
	}}, nil
}

func (s *RPCServer) DelCSP(ctx context.Context,
	req *proto.DelCSPReq) (*proto.DelCSPReply, error) {
	badReq := &proto.DelCSPReply{Code: 40000, Message: "Bad Param"}
	if !bson.IsObjectIdHex(req.UserId) {
		return badReq, nil
	}

	log.Debug("del csp with req: %v", req)
	err := s.cspManager.DelCSP(bson.ObjectIdHex(req.UserId), req.CspType)
	if err != nil {
		return &proto.DelCSPReply{Code: 50000, Message: err.Error()}, nil
	}

	return &proto.DelCSPReply{}, nil
}

package main

import (
	"context"
	"encoding/base64"
	"net"
	"time"

	"github.com/ICKelin/cframe/codec/proto"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/ICKelin/cframe/usercenter/models"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

type Server struct {
	addr        string
	userManager *models.UserManager
	authManager *models.AuthManager
}

func NewServer(addr string) *Server {
	return &Server{
		addr:        addr,
		userManager: models.GetUserManager(),
		authManager: models.GetAuthManager(),
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Info("listenning %v success", s.addr)
	srv := grpc.NewServer()
	proto.RegisterUserServiceServer(srv, s)
	return srv.Serve(listener)
}

func (s *Server) GetUserBySecret(ctx context.Context,
	req *proto.GetUserBySecretReq) (*proto.GetUserBySecretReply, error) {
	user, err := s.userManager.GetUserBySecret(req.Secret)
	if err != nil {
		log.Error("get user by secret %s fail: %v", req.Secret, err)
		return &proto.GetUserBySecretReply{Code: 50000, Message: err.Error()}, nil
	}

	log.Debug("getuser by secret req: %v", req)
	return &proto.GetUserBySecretReply{
		UserInfo: &proto.UserInfo{
			UserId:     user.Id.Hex(),
			UserName:   user.Username,
			UserSecret: user.Secret,
		},
	}, nil
}

func (s *Server) AddUser(ctx context.Context,
	req *proto.AddUserReq) (*proto.AddUserReply, error) {

	badReq := &proto.AddUserReply{Code: 40000, Message: "Bad Param"}
	if len(req.UserName) <= 0 ||
		len(req.Password) <= 0 ||
		len(req.Email) <= 0 {
		return badReq, nil
	}

	log.Debug("add user req: %v", req)

	// verify user
	exist, _ := s.userManager.GetUserByName(req.UserName)
	if exist != nil {
		return &proto.AddUserReply{Code: 50000, Message: "user exist"}, nil
	}

	exist, _ = s.userManager.GetUserByEmail(req.Email)
	if exist != nil {
		return &proto.AddUserReply{Code: 50000, Message: "email exist"}, nil
	}

	userInfo := &models.User{
		Username: req.UserName,
		Password: req.Password,
		Email:    req.Email,
		About:    req.About,
		Secret:   generateSecret(),
	}

	r, err := s.userManager.CreateUser(userInfo)
	if err != nil {
		return &proto.AddUserReply{Code: 50000, Message: err.Error()}, nil
	}
	return &proto.AddUserReply{
		UserInfo: &proto.UserInfo{
			UserId:     r.Id.Hex(),
			UserName:   r.Username,
			UserEmail:  r.Email,
			About:      r.About,
			UserSecret: r.Secret,
		},
	}, nil
}

func (s *Server) GetUserInfo(ctx context.Context,
	req *proto.GetUserInfoReq) (*proto.GetUserInfoReply, error) {
	log.Debug("get user req: %v", req)

	user, err := s.userManager.VerifyUser(req.UserName, req.Password)
	if err != nil {
		return &proto.GetUserInfoReply{Code: 50000, Message: err.Error()}, nil
	}
	return &proto.GetUserInfoReply{
		UserInfo: &proto.UserInfo{
			UserId:     user.Id.Hex(),
			UserName:   user.Username,
			UserSecret: user.Secret,
			UserEmail:  user.Email,
			CreatedAt:  user.CreatedAt,
		},
	}, nil
}

func (s *Server) Authorize(ctx context.Context,
	req *proto.AuthorizeReq) (*proto.AuthorizeReply, error) {
	log.Debug("auth req: %v", req)
	userInfo, err := s.userManager.VerifyUser(req.Username, req.Password)
	if err != nil {
		// TODO: verify by email
		return &proto.AuthorizeReply{Code: 50000, Message: err.Error()}, nil
	}

	authInfo := &models.Auth{
		UserId:    userInfo.Id,
		Token:     generateToken(),
		ExpiredIn: time.Now().Add(time.Hour * 1).Unix(),
	}

	r, err := s.authManager.AddAuth(authInfo)
	if err != nil {
		return &proto.AuthorizeReply{Code: 50000, Message: err.Error()}, nil
	}

	return &proto.AuthorizeReply{Data: &proto.AuthorizeReplyBody{
		UserId:    userInfo.Id.Hex(),
		Token:     r.Token,
		ExpiredIn: r.ExpiredIn,
	}}, nil

}

func (s *Server) GetUserByToken(ctx context.Context,
	req *proto.GetUserByTokenReq) (*proto.GetUserByTokenReply, error) {

	log.Debug("get user by token req: %v", req)
	authInfo, err := s.authManager.GetAuthByToken(req.Token)
	if err != nil {
		log.Error("get user by token fail: %v", err)
		return &proto.GetUserByTokenReply{Code: 50000, Message: err.Error()}, nil
	}

	userInfo, err := s.userManager.GetUserById(authInfo.UserId)
	if err != nil {
		log.Error("get user info fail: %v", err)
		return &proto.GetUserByTokenReply{Code: 50000, Message: err.Error()}, nil
	}

	return &proto.GetUserByTokenReply{Data: &proto.UserInfo{
		UserId:     userInfo.Id.Hex(),
		UserName:   userInfo.Username,
		UserEmail:  userInfo.Email,
		UserSecret: userInfo.Secret,
		About:      userInfo.About,
		CreatedAt:  userInfo.CreatedAt,
	}}, nil
}

func generateSecret() string {
	uniq := uuid.NewV4()
	return base64.StdEncoding.EncodeToString(uniq.Bytes())
}

func generateToken() string {
	uniq := uuid.NewV4()
	return base64.StdEncoding.EncodeToString(uniq.Bytes())
}

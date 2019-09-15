package controller

import (
	"encoding/json"
	"log"
	"net/http"
)

var (
	defaultAddr = ":10033"
)

type ServerConfig struct {
	Addr string `toml:"addr" json:"addr"`
}

type Server struct {
	addr    string
	nodeMgr *NodeManager
}

func NewServer(cfg ServerConfig, nm *NodeManager) *Server {
	if len(cfg.Addr) == 0 {
		cfg.Addr = defaultAddr
	}

	return &Server{
		nodeMgr: nm,
		addr:    cfg.Addr,
	}
}

func (s *Server) ListenAndServe() {
	log.Println("listening: ", s.addr)
	http.HandleFunc("/api/v1/nodes", s.getNodes)
	http.ListenAndServe(s.addr, nil)
}

func (s *Server) getNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.nodeMgr.GetNodes()
	b, _ := json.Marshal(nodes)
	w.Write(b)
}

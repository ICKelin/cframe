package controller

import (
	"encoding/json"
	"net/http"
)

func Main() {
}

type server struct {
	nodeMgr *NodeManager
}

func newServer() *server {
	return &server{
		nodeMgr: &NodeManager{},
	}
}

func (s *server) ListenAndServe() {
	http.HandleFunc("/api/v1/nodes", s.getNodes)
	http.ListenAndServe(":10033", nil)
}

func (s *server) getNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.nodeMgr.GetNodes()
	b, _ := json.Marshal(nodes)
	w.Write(b)
}

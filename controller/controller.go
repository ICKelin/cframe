package controller

import (
	"log"
)

func Main(confpath string) {
	cfg, err := ParseConfig(confpath)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(cfg)

	nodeManager := NewNodeManager(cfg.Nodes)
	s := NewServer(cfg.ServerConfig, nodeManager)
	s.ListenAndServe()
}

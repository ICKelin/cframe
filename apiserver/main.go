package main

import (
	"flag"
	"fmt"

	"github.com/ICKelin/cframe/pkg/edgemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)
	log.Info("%v", conf)

	// create etcd storage
	store := edgemanager.NewEtcdStorage(conf.Etcd)

	// create edge manager
	edgemanager.New(store)

	// create edge host manager
	edgemanager.NewEdgeHostManager(store)

	// create api server
	s := NewApiServer(conf.ApiAddr)
	s.Run()
}

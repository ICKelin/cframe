package main

import (
	"flag"
	"fmt"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
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
	log.Debug("%v", conf)

	// create etcd storage
	store := etcdstorage.NewEtcd(conf.Etcd)

	// create edge manager
	edgeManager := models.NewEdgeManager(store)

	// create route manager
	routeManager := models.NewRouteManager(store)

	// create namespace manager
	namespaceManager := models.NewNamespaceManager(store)

	// registry server for edge
	r := NewRegistryServer(conf.ListenAddr, edgeManager, routeManager, namespaceManager)

	// watch for edge delete/put
	// notify online edge
	go edgeManager.Watch(
		func(namespace string, edg *codec.Edge) {
			r.DelEdge(namespace, edg)
		},
		func(userId string, edg *codec.Edge) {
			r.ModifyEdge(namespace, edg)
		})

	// watch for route delete/put
	// notify online edge
	go routeManager.Watch(
		func(namespace string, route *codec.Route) {
			r.DelRoute(namespace, route)
		},
		func(namespace string, route *codec.Route) {
			r.AddRoute(namespace, route)
		},
	)
	r.ListenAndServe()
}

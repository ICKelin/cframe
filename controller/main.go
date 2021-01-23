package main

import (
	"flag"
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
	"github.com/ICKelin/cframe/pkg/database"
	"github.com/ICKelin/cframe/pkg/edgemanager"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/ICKelin/cframe/pkg/routemanager"
	"google.golang.org/grpc"
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

	cli, err := createUserServiceCli(conf.UserCenterAddr)
	if err != nil {
		log.Error("create user service fail: %v", err)
		return
	}

	// create etcd storage
	store := etcdstorage.NewEtcd(conf.Etcd)

	// create edge manager
	edgeManager := edgemanager.New(store)

	// create route manager
	routeManager := routemanager.New(store)

	// initial mongodb url
	database.NewModelManager(conf.MongoUrl, conf.DBName)

	rpcsrv := NewRPCServer(conf.RpcAddr)
	go rpcsrv.ListenAndServe()

	r := NewRegistryServer(conf.ListenAddr, cli)

	// watch for edge delete/put
	// tell registry edge change
	go edgeManager.Watch(
		func(edg *edgemanager.Edge) {
			r.DelEdge(edg)
		},
		func(edg *edgemanager.Edge) {
			r.ModifyEdge(edg)
		})

	// watch for route delete/put
	// tell registry route change
	go routeManager.Watch(
		func(route *routemanager.Route) {
			r.DelRoute(route)
		},
		func(route *routemanager.Route) {
			r.AddRoute(route)
		},
	)
	r.ListenAndServe()
}

func createUserServiceCli(remote string) (proto.UserServiceClient, error) {
	conn, err := grpc.Dial(remote, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	cli := proto.NewUserServiceClient(conn)
	return cli, nil
}

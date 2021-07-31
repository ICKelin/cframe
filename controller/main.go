package main

import (
	"flag"
	"fmt"
	"time"

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
	log.Debug("%v", conf)

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

	// rpc server api-service
	rpcsrv := NewRPCServer(conf.RpcAddr)
	go rpcsrv.ListenAndServe()

	// registry server for edge
	r := NewRegistryServer(conf.ListenAddr, cli)

	// watch for edge delete/put
	// notify online edge
	go edgeManager.Watch(
		func(userId string, edg *edgemanager.Edge) {
			r.DelEdge(userId, edg)
		},
		func(userId string, edg *edgemanager.Edge) {
			r.ModifyEdge(userId, edg)
		})

	// watch for route delete/put
	// notify online edge
	go routeManager.Watch(
		func(userId string, route *routemanager.Route) {
			r.DelRoute(userId, route)
		},
		func(userId string, route *routemanager.Route) {
			r.AddRoute(userId, route)
		},
	)
	r.ListenAndServe()
}

func createUserServiceCli(remote string) (proto.UserServiceClient, error) {
	for {
		conn, err := grpc.Dial(remote, grpc.WithInsecure(), grpc.WithTimeout(time.Second*10))
		if err != nil {
			log.Error("connect to user service fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		cli := proto.NewUserServiceClient(conn)
		return cli, nil
	}
}

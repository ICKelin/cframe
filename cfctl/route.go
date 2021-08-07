package main

import (
	"fmt"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

func delRoute(ns, name string, store *etcdstorage.Etcd) {
	routeMgr := models.NewRouteManager(store)
	err := routeMgr.DelRoute(ns, name)
	if err != nil {
		fmt.Printf("del route %s ret: %v", name, err)
		return
	}
	fmt.Printf("del route %s OK\n", name)
}

func addRoute(ns, name, listener, cidr string, store *etcdstorage.Etcd) {
	routeMgr := models.NewRouteManager(store)
	err := routeMgr.AddRoute(ns, &codec.Route{
		Name:    name,
		CIDR:    cidr,
		Nexthop: listener,
	})
	if err != nil {
		fmt.Printf("add route %s ret: %v", name, err)
		return
	}
	fmt.Printf("add route %s OK\n", name)
}

func listRoutes(ns string, store *etcdstorage.Etcd) {
	routeMgr := models.NewRouteManager(store)
	routes := routeMgr.GetRoutes(ns)

	fmt.Printf("\nroutes for %s namespace\n", ns)
	fmt.Printf("      %-20s %-25s %-15s\n", "Name", "Listener", "CIDR")
	fmt.Println("-----------------------------------------------------------")
	for i, r := range routes {
		fmt.Printf("%-5d %-20s %-25s %-15s\n", i+1, r.Name, r.Nexthop, r.CIDR)
	}
	fmt.Println("OK")
}

package main

import (
	"fmt"

	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

func listRoutes(ns string, store *etcdstorage.Etcd) {
	routeMgr := models.NewRouteManager(store)
	routeMgr.GetRoutes(ns)

	fmt.Println("edge list:")
	fmt.Printf("      %-15s %-25s %-15s\n", "Name", "Listener", "CIDR")
	fmt.Println("-----------------------------------------------------------")
	for i, edge := range edges {
		fmt.Printf("%-5d %-15s %-25s %-15s\n", i+1, edge.Name, edge.ListenAddr, edge.Cidr)
	}
}

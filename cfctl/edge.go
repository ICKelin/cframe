package main

import (
	"fmt"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
)

func addEdge(ns, edgeName, listenAddr, cidr string, store *etcdstorage.Etcd) {
	edgeMgr := models.NewEdgeManager(store)
	edgeMgr.AddEdge(ns, &codec.Edge{
		Name:       edgeName,
		Cidr:       cidr,
		ListenAddr: listenAddr,
	})
	fmt.Printf("create edge %s cidr %s OK\n", listenAddr, cidr)
}

func delEdge(ns, edgeName string, store *etcdstorage.Etcd) {
	edgeMgr := models.NewEdgeManager(store)
	edgeMgr.DelEdge(ns, edgeName)
	fmt.Printf("delete edge %s OK\n", edgeName)
}

func listEdges(ns string, store *etcdstorage.Etcd) {
	edgeMgr := models.NewEdgeManager(store)
	edges := edgeMgr.GetEdges(ns)

	fmt.Println("edge list:")
	fmt.Printf("      %-15s %-25s %-15s\n", "Name", "ListenAddress", "CIDR")
	fmt.Println("-----------------------------------------------------------")
	for i, edge := range edges {
		fmt.Printf("%-5d %-15s %-25s %-15s\n", i+1, edge.Name, edge.ListenAddr, edge.Cidr)
	}
}

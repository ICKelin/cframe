package main

import (
	"encoding/base64"
	"fmt"

	"github.com/ICKelin/cframe/controller/models"
	"github.com/ICKelin/cframe/pkg/etcdstorage"
	uuid "github.com/satori/go.uuid"
)

func addNamespace(name string, store *etcdstorage.Etcd) {
	uniq := uuid.NewV4()
	secret := base64.StdEncoding.EncodeToString(uniq.Bytes())
	namespaceMgr := models.NewNamespaceManager(store)
	err := namespaceMgr.AddNamespace(&models.Namespace{
		Name:   name,
		Secret: secret,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	nsInfo, err := namespaceMgr.GetNamespace(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("create namespace %s, secret %s OK\n", nsInfo.Name, nsInfo.Secret)
}

func delNamespace(name string, store *etcdstorage.Etcd) {
	namespaceMgr := models.NewNamespaceManager(store)
	err := namespaceMgr.DelNamespace(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("delete namespace %s OK\n", name)
}

func listNamespace(store *etcdstorage.Etcd) {
	namespaceMgr := models.NewNamespaceManager(store)
	nss := namespaceMgr.GetNamespaces()
	fmt.Println("namespace list:")
	fmt.Printf("      %-15s %-30s\n", "Name", "SecretKey")
	fmt.Println("--------------------------------------------")
	for i, ns := range nss {
		fmt.Printf("%-5d %-15s %-30s\n", i+1, ns.Name, ns.Secret)
	}
}

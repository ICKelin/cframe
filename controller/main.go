package main

import (
	"flag"
	"fmt"

	"github.com/ICKelin/cframe/controller/apiserver"
	"github.com/ICKelin/cframe/controller/edagemanager"
	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	log.Init("./log/controller.log", "debug", 3)
	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Info("%v", conf)

	// create etcd storage
	store := edagemanager.NewEtcdStorage(conf.Etcd)

	// create build in edages
	edageManager := edagemanager.New(store)
	for _, edageConf := range conf.BuildInEdages {
		edage := &edagemanager.Edage{
			Name:     edageConf.Name,
			HostAddr: edageConf.HostAddr,
			Cidr:     edageConf.Cidr,
		}

		log.Info("create build in edage %v", edage)
		if edageManager.VerifyCidr(edage.Cidr) == false {
			log.Info("create edage %v fail,conflict exist\n", edage)
			continue
		}
		edageManager.AddEdage(edage.Name, edage)
	}

	// create edage host manager
	edagemanager.NewEdageHostManager(store)

	// create api server
	s := apiserver.New(conf.ApiAddr)
	go s.Run()

	r := NewRegistryServer(conf.ListenAddr)
	r.ListenAndServe()
}

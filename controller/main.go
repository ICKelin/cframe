package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ICKelin/cframe/controller/edagemanager"
)

func main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	log.SetFlags(log.Lshortfile)
	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// create build in edages
	edageManager := edagemanager.New()
	for _, edageConf := range conf.BuildInEdages {
		edage := &edagemanager.Edage{
			Name:     edageConf.Name,
			HostAddr: edageConf.HostAddr,
			Cidr:     edageConf.Cidr,
		}

		log.Printf("create build in edage %v", edage)
		if edageManager.VerifyCidr(edage.Cidr) == false {
			log.Printf("[E] create edage %v fail,conflict exist\n", edage)
			continue
		}
		edageManager.AddEdage(edage.Name, edage)
	}

	r := NewRegistryServer(conf.ListenAddr)
	r.ListenAndServe()
}

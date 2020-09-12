package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ICKelin/cframe/edge/vpc"
	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	confpath := flag.String("c", "", "config file")
	flag.Parse()

	log.Init("log/edge.log", "debug", 3)

	cfg, err := ParseConfig(*confpath)
	if err != nil {
		fmt.Printf("parse config fali: %v\n", err)
		return
	}

	iface, err := NewInterface()
	if err != nil {
		log.Error("[E] new interface fail: ", err)
		return
	}

	defer iface.Close()
	iface.Up()

	// create VPC Instance
	accessKey := cfg.AccessKey
	secret := cfg.AccessSecret
	vpcInstance, err := vpc.GetVPCInstance(cfg.Type, accessKey, secret)
	if err != nil {
		log.Error("%v", err)
		return
	}

	// create cframe udp server
	s := NewServer(cfg.ListenAddr, iface, vpcInstance)

	// create registry to get connect to controller
	reg := NewRegistry(cfg.Controller, cfg.Name, s)
	go func() {
		err := reg.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}()

	s.SetRegistry(reg)
	s.ListenAndServe()
}

package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	confpath := flag.String("c", "", "config file")
	flag.Parse()

	cfg, err := ParseConfig(*confpath)
	if err != nil {
		fmt.Printf("parse config fali: %v\n", err)
		return
	}
	log.Init(cfg.Log.Path, cfg.Log.Level, cfg.Log.Days)

	iface, err := NewInterface()
	if err != nil {
		log.Error("[E] new interface fail: ", err)
		return
	}

	defer iface.Close()
	iface.Up()

	// create cframe udp server
	// just hard code listen address once without env var
	lisAddr := ":58423"
	lis := os.Getenv("listen")
	if len(lis) > 0 {
		lisAddr = lis
	}
	s := NewServer(lisAddr, cfg.SecretKey, iface, nil)

	// create registry to get connect to controller
	// just hard code controller address once without env var
	ctrlAddr := "demo.notr.tech:58422"
	ctrl := os.Getenv("controller")
	if len(ctrl) > 0 {
		ctrlAddr = ctrl
	}

	// it is our secret
	// read from env firstly
	// if empty, use configuration
	secret := os.Getenv("secret")
	if len(secret) <= 0 {
		secret = cfg.SecretKey
	}

	reg := NewRegistry(ctrlAddr, cfg.SecretKey, s)
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

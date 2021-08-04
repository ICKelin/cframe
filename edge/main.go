package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) == 0 {
		logLevel = "debug"
	}
	log.Init("log/edge.log", logLevel, 3)

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

	ipport := strings.Split(lisAddr, ":")
	if len(ipport) != 2 {
		log.Error("invalid listen addr %s, should be ip:p port format", lisAddr)
		return
	}

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
		log.Error("invalid secret")
		return
	}

	ns := os.Getenv("namespace")
	if len(ns) <= 0 {
		log.Error("invalid namespace")
		return
	}

	s := NewServer(lisAddr, secret, iface)

	reg := NewRegistry(ctrlAddr, ns, secret, ipport[1], s)
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

package main

import (
	"fmt"
	"os"

	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	log.Init("log/edge.log", "debug", 3)

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

	s := NewServer(lisAddr, secret, iface, nil)

	reg := NewRegistry(ctrlAddr, secret, s)
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

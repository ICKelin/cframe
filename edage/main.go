package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	confpath := flag.String("c", "", "config file")
	flag.Parse()

	cfg, err := ParseConfig(*confpath)
	if err != nil {
		fmt.Printf("parse config fali: %v\n", err)
		return
	}

	iface, err := NewInterface()
	if err != nil {
		log.Println("[E] new interface fail: ", err)
		return
	}

	defer iface.Close()
	iface.Up()

	reg := NewRegistry(cfg.Controller, cfg.Local.Addr, cfg.Local.CIDR)
	go func() {
		err := reg.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}()

	s := NewServer(cfg.Local.Addr, cfg.Peers, iface)
	s.ListenAndServe()
}

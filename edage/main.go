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

	log.SetFlags(log.Lshortfile)

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

	s := NewServer(cfg.ListenAddr, iface)

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

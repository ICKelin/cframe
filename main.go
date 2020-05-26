package main

import (
	"flag"
	"log"
)

func main() {
	laddr := flag.String("l", ":58422", "local udp address")
	remote := flag.String("r", "172.18.171.245:58422", "remote addr")
	cidr := flag.String("cidr", "192.168.10.0/24", "remote cidr")

	flag.Parse()

	nodes := []*Node{
		{
			Addr: *remote,
			CIDR: *cidr,
		},
	}

	iface, err := NewInterface()
	if err != nil {
		log.Println("[E] new interface fail: ", err)
		return
	}

	defer iface.Close()
	iface.Up()

	s := NewServer(*laddr, nodes, iface)
	s.ListenAndServe()
}

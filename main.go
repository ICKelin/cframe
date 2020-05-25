package main

import "flag"

func main() {
	laddr := flag.String("l", ":58422", "local udp address")
	remote := flag.String("r", "172.18.171.245:58422", "remote addr")
	flag.Parse()

	nodes := []*Node{
		{
			Addr: *remote,
		},
	}

	s := NewServer(*laddr, nodes)
	s.ListenAndServe()
}

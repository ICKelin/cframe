package cframe

import "flag"

func Main() {
	flgServer := flag.String("s", "", "server address")
	flgLan := flag.String("lan", "", "lan address")
	flgMask := flag.Int("mask", 32, "mask")
	flag.Parse()

	c := NewClient(*flgServer, *flgLan, int32(*flgMask))
	c.Run()

}

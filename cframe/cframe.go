package cframe

import (
	"flag"
	"fmt"
	"log"
)

func Main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	cfg, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println(cfg)

	ctrl := NewCtrlClient(cfg.CtrlConfig)
	c := NewClient(cfg.ClientConfig, ctrl)
	c.Run()
}

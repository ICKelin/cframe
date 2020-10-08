package main

import (
	"flag"
	"fmt"

	log "github.com/ICKelin/cframe/pkg/logs"
)

func main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)
	log.Info("%v", conf)

	// create api server
	s := NewApiServer(conf.ApiAddr, conf.UserCenterAddr, conf.CtrlAddr)
	s.Run()
}

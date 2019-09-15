package main

import (
	"flag"

	"github.com/ICKelin/cframe/controller"
)

func main() {
	flgConf := flag.String("c", "", "conf path")
	flag.Parse()
	controller.Main(*flgConf)
}

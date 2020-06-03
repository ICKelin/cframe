package main

import "flag"

func main() {
	local := flag.String("l", "", "local address")
	flag.Parse()

	r := NewRegistryServer(*local)
	r.ListenAndServe()
}

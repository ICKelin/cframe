package main

import "testing"

func TestRegister(t *testing.T) {
	r := NewRegistry("127.0.0.1:1234", "127.0.0.1:1235", "192.168.10.0/24")
	r.run()
}

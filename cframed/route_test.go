package cframed

import (
	"net"
	"testing"
)

func TestRoute(t *testing.T) {
	r := NewRouter()
	r.NewItem(net.ParseIP("192.168.0.1"), 24, &net.TCPConn{})
	conn := r.NextHop("192.168.0.20")
	if conn == nil {
		t.Error("expect not nil")
		return
	}

	conn = r.NextHop("192.168.0.255")
	if conn == nil {
		t.Error("expect not nil")
		return
	}

	conn = r.NextHop("192.168.1.0")
	if conn != nil {
		t.Error("expect nil")
		return
	}

	conn = r.NextHop("192.168.1.255")
	if conn != nil {
		t.Error("expect nil")
		return
	}
}

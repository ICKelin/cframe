package cframe

import (
	"testing"
)

func TestClient(t *testing.T) {
	c := NewClient("120.25.214.63:10222", "192.168.0.1", 24)
	c.Run()
}

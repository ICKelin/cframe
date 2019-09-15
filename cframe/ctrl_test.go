package cframe

import (
	"encoding/json"
	"testing"
)

func TestCtrl(t *testing.T) {
	c := NewCtrlClient(CtrlConfig{
		Scheme: "http://",
		Addr:   "127.0.0.1:10033",
	})

	nodes, err := c.GetNodes()
	if err != nil {
		t.Error(err)
		return
	}

	b, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(b))
}

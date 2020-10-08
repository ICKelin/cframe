package edgemanager

import (
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
)

const (
	StatusOffline = iota
	StatusOnline
)

type Edge struct {
	Name       string        `json:"name"`
	Comment    string        `json:"comment"`
	Cidr       string        `json:"cidr"`
	ListenAddr string        `json:"listen_addr"`
	Type       proto.CSPType `json:"type"`
}

func (e *Edge) String() string {
	return fmt.Sprintf("name: %s type %s listenaddr: %s cidr: %s",
		e.Name, e.Type, e.ListenAddr, e.Cidr)
}

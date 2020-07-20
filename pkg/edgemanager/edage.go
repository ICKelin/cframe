package edgemanager

import (
	"fmt"
)

const (
	StatusOffline = iota
	StatusOnline
)

type Edge struct {
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	Cidr     string `json:"cidr"`
	HostAddr string `json:"host_addr"`
	Status   int    `json:"status"`
}

func (e *Edge) String() string {
	return fmt.Sprintf("name: %s hostaddr: %s cidr: %s",
		e.Name, e.HostAddr, e.Cidr)
}

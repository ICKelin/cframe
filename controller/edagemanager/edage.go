package edagemanager

import (
	"fmt"
)

const (
	StatusOffline = iota
	StatusOnline
)

type Edage struct {
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	Cidr     string `json:"cidr"`
	HostAddr string `json:"host_addr"`
	Status   int    `json:"status"`
}

func (e *Edage) String() string {
	return fmt.Sprintf("name: %s hostaddr: %s cidr: %s",
		e.Name, e.HostAddr, e.Cidr)
}

func (e *Edage) IsOnline() bool {
	return e.GetStatus() == StatusOnline
}

func (e *Edage) IsOffline() bool {
	return e.GetStatus() == StatusOffline
}

func (e *Edage) GetStatus() int {
	return e.Status
}

func (e *Edage) SetOnline() {
	e.SetStatus(StatusOnline)
}

func (e *Edage) SetOffline() {
	e.SetStatus(StatusOffline)
}

func (e *Edage) SetStatus(status int) {
	e.Status = status
}

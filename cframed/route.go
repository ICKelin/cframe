package cframed

import (
	"net"
	"sync"
)

type Route struct {
	sync.RWMutex
	tableEntry [32][]*RouteItem
}

type RouteItem struct {
	ip      net.IP
	mask    int32
	network int32
	conn    net.Conn
}

func NewRouter() *Route {
	r := &Route{}
	for i := 0; i < 32; i++ {
		r.tableEntry[i] = make([]*RouteItem, 0)
	}

	return r
}

func (r *Route) NewItem(ip net.IP, mask int32, nexthop net.Conn) {
	m := (1 << 32) - (1 << (32 - uint32(mask)))
	network := ipaddr(ip.String()) & int32(m)
	item := &RouteItem{
		ip:      ip,
		mask:    int32(m),
		network: network,
		conn:    nexthop,
	}

	r.Lock()
	defer r.Unlock()
	r.tableEntry[mask] = append(r.tableEntry[mask], item)
	// ip ro add ip/mask dev tun0
}

func (r *Route) RemoveItem(ip net.IP, mask int32) {
	r.Lock()
	defer r.Unlock()

	// ip ro del ip/mask dev tun0

	entry := r.tableEntry[mask]
	for i, e := range entry {
		if e.ip.String() == ip.String() && e.mask == mask {
			entry = append(entry[:i], entry[i+1:]...)
			return
		}
	}
}

func (r *Route) NextHop(ip string) net.Conn {
	r.RLock()
	defer r.RUnlock()

	for i := 31; i >= 0; i-- {
		entry := r.tableEntry[i]
		for _, e := range entry {
			iip := ipaddr(ip)
			if (iip & e.mask) == e.network {
				return e.conn
			}
		}
	}
	return nil
}

func ipaddr(sip string) int32 {
	ip := net.ParseIP(sip)
	ipv4 := ip.To4()
	return (int32(ipv4[0]) << 24) + (int32(ipv4[1]) << 16) + (int32(ipv4[2]) << 8) + int32(ipv4[3])
}

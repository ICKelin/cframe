// proto.go defines transfer protocol between
// edge and controller
// includes the following sections:
//  1. edge register
//  2. edge online
//  3. edge offline
//  @ICKelin 2020.07.05

package codec

import "encoding/json"

// edge register req
type RegisterReq struct {
	// edge name
	Name string
}

// edge host information
type Host struct {
	HostAddr string
	Cidr     string
}

// reply for edge register req
type RegisterReply struct {
	OnlineHost []*Host
}

func (r *RegisterReply) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// broadcast edge online
// once edge register success
// controller will broadcast edge online msg
// to all online edges.
type BroadcastOnlineMsg struct {
	// onlined edge host address(udp://ip:port)
	HostAddr string

	// offline edge network subnet(192.168.10.0/24)
	Cidr string
}

// broadcase edge offline
// once edge keepalive faile
// controller will broadcase edge offline msg
// to all online edges
type BroadcastOfflineMsg struct {
	// offlined edge host address
	HostAddr string

	// offlined edge network subnet
	Cidr string
}

// edge report host
// edge report host information
// to controller
// controller get topology
//  host1                    host1'
//       \                  /
//        edge1 ---- edge2
//       /                  \
//  host2                    host2'
//
type ReportEdgeHost struct {
	HostIPs []string
}

type Heartbeat struct{}

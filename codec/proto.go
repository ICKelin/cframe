// proto.go defines transfer protocol between
// edage and controller
// includes the following sections:
//  1. edage register
//  2. edage online
//  3. edage offline
//  @ICKelin 2020.07.05

package codec

import "encoding/json"

// edage register req
type RegisterReq struct {
	// edage name
	Name string
}

// edage host information
type Host struct {
	HostAddr string
	Cidr     string
}

// reply for edage register req
type RegisterReply struct {
	OnlineHost []*Host
}

func (r *RegisterReply) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// broadcast edage online
// once edage register success
// controller will broadcast edage online msg
// to all online edages.
type BroadcastOnlineMsg struct {
	// onlined edage host address(udp://ip:port)
	HostAddr string

	// offline edage network subnet(192.168.10.0/24)
	Cidr string
}

// broadcase edage offline
// once edage keepalive faile
// controller will broadcase edage offline msg
// to all online edages
type BroadcastOfflineMsg struct {
	// offlined edage host address
	HostAddr string

	// offlined edage network subnet
	Cidr string
}

type Heartbeat struct{}

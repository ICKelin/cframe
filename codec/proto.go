// proto.go defines transfer protocol between
// edge and controller
// includes the following sections:
//  1. edge register
//  2. edge online
//  3. edge offline
//  @ICKelin 2020.07.05

package codec

import (
	"encoding/json"
	"fmt"
)

type CSPType int

const (
	CSP_TYPE_NONE = iota
	CSP_TYPE_ALI
	CSP_TYPE_AWS
)

type Route struct {
	CIDR    string
	Nexthop string
	Name    string
}

type Edge struct {
	Name       string  `json:"name"`
	Cidr       string  `json:"cidr"`
	ListenAddr string  `json:"listen_addr"`
	Type       CSPType `json:"type"`
}

// edge register req
type RegisterReq struct {
	Namespace string
	SecretKey string
	PublicIP  string
}

func (e *Edge) String() string {
	return fmt.Sprintf("listen %s, local cidr %s", e.ListenAddr, e.Cidr)
}

type CSPInfo struct {
	CspType      CSPType
	AccessKey    string
	AccessSecret string
}

// reply for edge register req
type RegisterReply struct {
	EdgeList []*Edge
	CSPInfo  *CSPInfo
	Routes   []*Route
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
	// onlined edge listen address(udp://ip:port)
	ListenAddr string

	// offline edge network subnet(192.168.10.0/24)
	Cidr string
}

// broadcase edge offline
// once edge keepalive faile
// controller will broadcase edge offline msg
// to all online edges
type BroadcastOfflineMsg struct {
	// offlined edge host address
	ListenAddr string

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
type ReportMsg struct {
	Timestamp  int64
	CPU        int32
	Mem        int32
	TrafficIn  int64
	TrafficOut int64
	Error      []string
}

type Heartbeat struct{}

// controller deploy route added to edges
type AddRouteMsg struct {
	// dst cidr
	Cidr string
	// next hop edge listen address
	// ip:port
	Nexthop string
}

// controller deploy route deleted to edges
type DelRouteMsg AddRouteMsg

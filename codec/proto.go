package codec

import "encoding/json"

// 节点注册请求
type RegisterReq struct {
	// edage name
	Name string
}

type Host struct {
	HostAddr string
	Cidr     string
}

// 节点注册响应
type RegisterReply struct {
	OnlineHost []*Host
}

func (r *RegisterReply) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// 广播节点上线消息
type BroadcastOnlineMsg struct {
	// 对端监听地址
	HostAddr string

	// 容器网段
	Cidr string
}

// 广播节点下线消息
type BroadcastOfflineMsg struct {
	// 对端监听地址
	HostAddr string

	// 容器网段
	Cidr string
}

type Heartbeat struct{}

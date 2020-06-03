package codec

import "encoding/json"

// 节点注册请求
type RegisterReq struct {
	// 主机监听的地址
	HostAddr string

	// 容器网段
	ContainerCidr string
}

type Host struct {
	HostAddr      string
	ContainerCidr string
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
	ContainerCidr string
}

// 广播节点下线消息
type BroadcastOfflineMsg struct {
	// 对端监听地址
	HostAddr string

	// 容器网段
	ContainerCidr string
}

type Heartbeat struct{}

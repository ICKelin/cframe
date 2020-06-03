package codec

// 节点注册请求
type RegisterReq struct {
	// 主机监听的地址
	HostAddr string

	// 容器网段
	ContainerCidr string
}

// 节点注册响应
type RegisterReply struct{}

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
}

type Heartbeat struct{}

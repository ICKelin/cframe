package controller

type Node struct {
	CIDR string `json:"cidr" toml:"cidr"`
	Name string `json:"name" toml:"name"`
}

type NodeManager struct {
	nodes []*Node
}

func NewNodeManager(nodes []*Node) *NodeManager {
	return &NodeManager{
		nodes: nodes,
	}
}

func (n *NodeManager) GetNodes() []*Node {
	return n.nodes
}

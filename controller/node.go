package controller

type Node struct {
	CIDR string `json:"cidr"`
	Name string `json:"name"`
}

type NodeManager struct {
	nodes []*Node
}

func (n *NodeManager) GetNodes() []*Node {
	return n.nodes
}

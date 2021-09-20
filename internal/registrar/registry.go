package registrar

type NodeRegistry struct {
	LocalNode     *Node
	ExternalNodes []*Node
}

func NewNodeRegistry(local *Node, external []*Node) *NodeRegistry {
	n := NodeRegistry{
		LocalNode:     local,
		ExternalNodes: external,
	}

	return &n
}

package courier

type NodeStorer interface {
	GetNodes([]string) ([]*node, error)
	AddNode(*node) error
	RemoveNode(*node) error
}

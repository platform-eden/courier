package registrar

type NodeStorer interface {
	GetNodes([]string) ([]*Node, error)
	AddNode(*Node) error
	RemoveNode(*Node) error
}

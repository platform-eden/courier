package courier

// interface that does the interactions with a Node Store
type Observer interface {
	Observe() (chan []Node, error)
	AddNode(*Node) error
	RemoveNode(*Node) error
}

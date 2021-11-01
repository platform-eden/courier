package courier

import "github.com/platform-edn/courier/node"

// interface that does the interactions with a Node Store
type Observer interface {
	Observe() (chan []node.Node, error)
	AddNode(*node.Node) error
	RemoveNode(*node.Node) error
}

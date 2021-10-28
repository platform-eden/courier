package observe

import "github.com/platform-edn/courier/node"

// interface that does the interactions with a Node Store
type NodeStorer interface {
	GetSubscribers(...string) ([]*node.Node, error)
	AddNode(*node.Node) error
	RemoveNode(*node.Node) error
}

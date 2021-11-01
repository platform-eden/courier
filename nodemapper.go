package courier

import "github.com/platform-edn/courier/node"

type NodeMapper interface {
	Node(string) (node.Node, bool)
	Update(...node.Node)
	Add(node.Node)
	Remove(string)
	Length() int
}

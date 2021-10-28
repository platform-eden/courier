package node

type NodeMapper interface {
	Node(string) (Node, bool)
	Add(Node)
	Remove(string)
	Length() int
}

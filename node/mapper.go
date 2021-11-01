package node

type NodeMapper interface {
	Node(string) (Node, bool)
	Update(...Node)
	Add(Node)
	Remove(string)
	Length() int
}

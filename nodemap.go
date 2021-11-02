package courier

import (
	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type NodeMapper interface {
	Node(string) (node.Node, bool)
	Nodes() map[string]node.Node
	Update(...node.Node)
	Add(node.Node)
	Remove(string)
	Length() int
}

type NodeMap struct {
	nodes map[string]node.Node
	lock  lock.Locker
}

func NewNodeMap(nodes ...node.Node) *NodeMap {
	nm := NodeMap{
		nodes: map[string]node.Node{},
		lock:  lock.NewTicketLock(),
	}

	for _, n := range nodes {
		nm.Add(n)
	}

	return &nm
}

func (nm *NodeMap) Node(id string) (node.Node, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
}

func (nm *NodeMap) Nodes() map[string]node.Node {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return nm.nodes
}

func (nm *NodeMap) Add(n node.Node) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.Id] = n
}

func (nm *NodeMap) Remove(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *NodeMap) Update(nodes ...node.Node) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	new := map[string]node.Node{}

	for _, n := range nodes {

		new[n.Id] = n
	}

	nm.nodes = new
}

func (nm *NodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

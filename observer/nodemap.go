package observer

import (
	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type NodeMap struct {
	nodes map[string]node.Node
	lock  lock.Locker
}

func NewNodeMap() *NodeMap {
	b := NodeMap{
		nodes: map[string]node.Node{},
		lock:  lock.NewTicketLock(),
	}

	return &b
}

func (b *NodeMap) Node(id string) (node.Node, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	n, exist := b.nodes[id]

	return n, exist
}

func (b *NodeMap) Nodes() map[string]node.Node {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.nodes
}

func (b *NodeMap) Update(nmap map[string]node.Node) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nodes = nmap
}

func (b *NodeMap) AddNode(n node.Node) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nodes[n.Id] = n
}

func (b *NodeMap) RemoveNode(id string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	delete(b.nodes, id)
}

func (b *NodeMap) Length() int {
	b.lock.Lock()
	defer b.lock.Unlock()

	return len(b.nodes)
}

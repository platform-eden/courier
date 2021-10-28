package observe

import (
	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type NodeMap struct {
	nodes map[string]node.Node
	lock  lock.Locker
}

func NewNodeMap() *NodeMap {
	nm := NodeMap{
		nodes: map[string]node.Node{},
		lock:  lock.NewTicketLock(),
	}

	return &nm
}

func (nm *NodeMap) Node(id string) (node.Node, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
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

func (nm *NodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

package registry

import "github.com/platform-edn/courier/pkg/lock"

type NodeMap struct {
	nodes map[string]Node
	lock  lock.Locker
}

func NewNodeMap(nodes ...Node) *NodeMap {
	nm := NodeMap{
		nodes: map[string]Node{},
		lock:  lock.NewTicketLock(),
	}

	for _, n := range nodes {
		nm.addNode(n)
	}

	return &nm
}

func (nm *NodeMap) Node(id string) (Node, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
}

func (nm *NodeMap) Nodes() map[string]Node {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return nm.nodes
}

func (nm *NodeMap) addNode(n Node) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.Id] = n
}

func (nm *NodeMap) removeNode(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *NodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

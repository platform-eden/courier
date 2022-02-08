package registry

import "github.com/platform-edn/courier/pkg/lock"

type NodeMap struct {
	Nodes map[string]Node
	lock.Locker
}

func NewNodeMap() *NodeMap {
	nm := NodeMap{
		Nodes:  map[string]Node{},
		Locker: lock.NewTicketLock(),
	}

	return &nm
}

func (nm *NodeMap) AddNode(n Node) {
	nm.Lock()
	defer nm.Unlock()

	nm.Nodes[n.Id] = n
}

func (nm *NodeMap) RemoveNode(id string) {
	nm.Lock()
	defer nm.Unlock()

	delete(nm.Nodes, id)
}

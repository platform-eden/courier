package courier

type NodeMap struct {
	nodes map[string]Node
	lock  Locker
}

func NewNodeMap(nodes ...Node) *NodeMap {
	nm := NodeMap{
		nodes: map[string]Node{},
		lock:  NewTicketLock(),
	}

	for _, n := range nodes {
		nm.Add(n)
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

func (nm *NodeMap) Add(n Node) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.id] = n
}

func (nm *NodeMap) Remove(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *NodeMap) Update(nodes ...Node) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	new := map[string]Node{}

	for _, n := range nodes {

		new[n.id] = n
	}

	nm.nodes = new
}

func (nm *NodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

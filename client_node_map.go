package courier

type ClientNodeMapper interface {
	Node(string) (clientNode, bool)
	Add(clientNode)
	Remove(string)
	Length() int
}

type clientNodeMap struct {
	nodes map[string]clientNode
	lock  Locker
}

func newClientNodeMap() *clientNodeMap {
	nm := clientNodeMap{
		nodes: map[string]clientNode{},
		lock:  NewTicketLock(),
	}

	return &nm
}

func (nm *clientNodeMap) Node(id string) (clientNode, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
}

func (nm *clientNodeMap) Add(n clientNode) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.id] = n
}

func (nm *clientNodeMap) Remove(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *clientNodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

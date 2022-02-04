package client

import (
	"log"

	"github.com/platform-edn/courier/pkg/lock"
)

type clientNodeMap struct {
	nodes map[string]clientNode
	lock  lock.Locker
}

func newClientNodeMap() *clientNodeMap {
	nm := clientNodeMap{
		nodes: map[string]clientNode{},
		lock:  lock.NewTicketLock(),
	}

	return &nm
}

func (nm *clientNodeMap) Node(id string) (clientNode, bool) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	n, exist := nm.nodes[id]

	return n, exist
}

func (nm *clientNodeMap) AddClientNode(n clientNode) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	nm.nodes[n.Id] = n
}

func (nm *clientNodeMap) RemoveClientNode(id string) {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	delete(nm.nodes, id)
}

func (nm *clientNodeMap) GenerateClientNodes(in <-chan string) <-chan clientNode {
	out := make(chan clientNode)
	go func() {
		for id := range in {
			n, exist := nm.Node(id)
			if !exist {
				log.Printf("node %s does not exist in nodemap - skipping", id)
				continue
			}

			out <- n
		}
		close(out)
	}()

	return out
}

func (nm *clientNodeMap) Length() int {
	nm.lock.Lock()
	defer nm.lock.Unlock()

	return len(nm.nodes)
}

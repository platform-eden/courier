package observer

import (
	"log"
	"time"

	"github.com/platform-edn/courier/node"
)

type Observer interface {
	ObserverInterval() time.Duration
	FailedConnectionChannel() chan node.Node
	GetCourierNodes() ([]*node.Node, error)
	TotalBlackListedNodes() int
	BlackListedNodes() map[string]node.Node
	UpdateBlackListedNodes(map[string]node.Node)
	AddBlackListNode(node.Node)
	CurrentNodes() map[string]node.Node
	RemoveCurrentNode(string)
	UpdateCurrentNodes(map[string]node.Node)
	SendNodes(map[string]node.Node)
}

// Starts a Goroutine that will begin comparing current nodes and what nodes the NodeStore has.  If the NodeStore updates,
// it sends a new map of Nodes to each Node Channel listening to the Observer.
func Observe(observer Observer) {
	for {
		select {
		case <-time.After(observer.ObserverInterval()):
			AttemptUpdatingNodes(observer)

		case n := <-observer.FailedConnectionChannel():
			observer.AddBlackListNode(n)
			observer.RemoveCurrentNode(n.Id)
		}
	}
}

// AttemptUpdatingNodes gets nodes from a node store, removes any blacklisted nodes, and sends updated nodes through a channel if there are any new ones or any removed
func AttemptUpdatingNodes(observer Observer) {
	nodes, err := observer.GetCourierNodes()
	if err != nil {
		log.Printf("could not observe Courier nodes: %s", err)
		return
	}

	nodes = CompareBlackListNodes(nodes, observer)
	CompareCurrentNodes(nodes, observer)
}

// CompareBlackListNodes checks if a list of nodes contains any blacklisted nodes.  If a node is blacklisted, it will now be added
// to the returned list of nodes.  This also returns an updated list of blacklisted nodes in case a currently blacklisted node is removed
// from the courier system.
func CompareBlackListNodes(nodes []*node.Node, observer Observer) []*node.Node {
	if observer.TotalBlackListedNodes() == 0 {
		return nodes
	}

	blacklist := observer.BlackListedNodes()
	nl := []*node.Node{}
	bl := map[string]node.Node{}

	for _, n := range nodes {
		_, exist := blacklist[n.Id]
		if exist {
			bl[n.Id] = *n
			log.Printf("Node %s is currently blacklisted - skipping node", n.Id)
		} else {
			nl = append(nl, n)
		}
	}

	observer.UpdateBlackListedNodes(blacklist)

	return nl
}

// compares the Nodes returned from the NodeStore with the current Nodes in the service.
// If there are differences, this will return true with an updated map of Nodes.
func CompareCurrentNodes(new []*node.Node, observer Observer) {
	current := observer.CurrentNodes()
	updated := map[string]node.Node{}
	update := false

	for _, n := range new {
		updated[n.Id] = *n

		_, exist := current[n.Id]
		if exist {
			delete(current, n.Id)
		} else {
			update = true
		}
	}

	if len(current) != 0 || update {
		observer.UpdateCurrentNodes(updated)
		observer.SendNodes(updated)
	}
}

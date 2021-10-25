package observer

import (
	"time"

	"github.com/platform-edn/courier/node"
)

type Observer interface {
	ObserverInterval() time.Duration
	AttemptUpdatingNodes()
	NodeChannel() chan (map[string]node.Node)
	FailedConnectionChannel() chan node.Node
	BlackListNode(node.Node)
}

// Starts a Goroutine that will begin comparing current nodes and what nodes the NodeStore has.  If the NodeStore updates,
// it sends a new map of Nodes to each Node Channel listening to the Observer.
func observe(o Observer) {
	for {
		select {
		case <-time.After(o.ObserverInterval()):
			o.AttemptUpdatingNodes()

		case n := <-o.FailedConnectionChannel():
			o.BlackListNode(n)
		}
	}
}

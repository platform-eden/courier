package observer

import (
	"log"
	"time"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type StoreObserver struct {
	store             NodeStorer
	observeInterval   time.Duration
	currentNodes      *NodeMap
	blackListedNodes  *NodeMap
	nodeChannels      []chan (map[string]node.Node)
	failedConnections chan node.Node
	subjects          []string
	lock              lock.Locker
}

// Returns a paused StoreObserver
func NewStoreObserver(store NodeStorer, interval time.Duration, subjects []string) *StoreObserver {
	s := StoreObserver{
		store:             store,
		observeInterval:   interval,
		currentNodes:      NewNodeMap(),
		blackListedNodes:  NewNodeMap(),
		nodeChannels:      []chan (map[string]node.Node){},
		failedConnections: make(chan node.Node),
		subjects:          subjects,
		lock:              lock.NewTicketLock(),
	}
	s.observe()
	return &s
}

// NodeChannel adds a channel to the StoreObserver that will receive a map of Nodes when the NodeStore has updated Nodes and returns it
func (s *StoreObserver) NodeChannel() chan (map[string]node.Node) {
	s.lock.Lock()
	defer s.lock.Unlock()

	channel := make(chan map[string]node.Node)
	s.nodeChannels = append(s.nodeChannels, channel)

	return channel
}

func (s *StoreObserver) FailedConnectionChannel() chan node.Node {
	return s.failedConnections
}

// Starts a Goroutine that will begin comparing current nodes and what nodes the NodeStore has.  If the NodeStore updates,
// it sends a new map of Nodes to each Node Channel listening to the Observer.
func (s *StoreObserver) observe() {
	go func() {
		for {
			select {
			case <-time.After(s.observeInterval):
				s.attemptUpdatingNodes()

			case n := <-s.failedConnections:
				s.blackListedNodes.AddNode(n)
				s.currentNodes.RemoveNode(n.Id)
			}
		}
	}()
}

func (s *StoreObserver) attemptUpdatingNodes() {
	nodes, err := s.store.GetSubscribers(s.subjects...)
	if err != nil {
		log.Printf("could not observe store nodes: %s", err)
		return
	}

	if s.blackListedNodes.Length() > 0 {
		var blacklist map[string]node.Node

		nodes, blacklist = compareBlackListNodes(nodes, s.blackListedNodes.Nodes())
		s.blackListedNodes.Update(blacklist)
	}

	current, updated := comparePotentialNodes(nodes, s.currentNodes.Nodes())
	if updated {
		s.currentNodes.Update(current)
		sendUpdatedNodes(current, s.nodeChannels)
	}
}

// CompareBlackListNodes checks if a list of nodes contains any blacklisted nodes.  If a node is blacklisted, it will now be added
// to the returned list of nodes.  This also returns an updated list of blacklisted nodes in case a currently blacklisted node is removed
// from the courier system.
func compareBlackListNodes(nodes []*node.Node, blacklist map[string]node.Node) ([]*node.Node, map[string]node.Node) {
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

	return nl, bl
}

// compares the Nodes returned from the NodeStore with the current Nodes in the service.
// If there are differences, this will return true with an updated map of Nodes.
func comparePotentialNodes(potential []*node.Node, expired map[string]node.Node) (map[string]node.Node, bool) {
	current := map[string]node.Node{}
	new := map[string]*node.Node{}
	updated := false

	for _, n := range potential {
		_, exist := expired[n.Id]

		if exist {
			current[n.Id] = *n
			delete(expired, n.Id)
		} else {
			_, exist := new[n.Id]
			if !exist {
				current[n.Id] = *n
				new[n.Id] = n
				updated = true
			}
		}
	}

	if len(expired) != 0 {
		updated = true
	}

	return current, updated
}

// SendUpdatedNodes takes a list of channels and sends a map of nodes with their id as the key
func sendUpdatedNodes(current map[string]node.Node, channels []chan (map[string]node.Node)) {
	for _, channel := range channels {
		channel <- current
	}
}

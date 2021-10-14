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
	currentNodes      map[string]node.Node
	blackListedNodes  map[string]node.Node
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
		currentNodes:      map[string]node.Node{},
		blackListedNodes:  map[string]node.Node{},
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
			timer := time.NewTimer(s.observeInterval)

			<-timer.C

			nodes, err := s.store.GetSubscribers(s.subjects...)
			if err != nil {
				log.Printf("%s could not observe store nodes", time.Now().String())
				continue
			}

			current, updated := compareNodes(nodes, s.currentNodes)

			if updated {
				s.currentNodes = current
				s.lock.Lock()
				for _, channel := range s.nodeChannels {
					channel <- current
				}
				s.lock.Unlock()
			}
		}
	}()
}

// compares the Nodes returned from the NodeStore with the current Nodes in the service.
// If there are differences, this will return true with an updated map of Nodes.
func compareNodes(potential []*node.Node, expired map[string]node.Node) (map[string]node.Node, bool) {
	current := map[string]node.Node{}
	new := map[string]*node.Node{}
	updated := false

	for _, node := range potential {
		_, ok := expired[node.Id]

		if ok {
			current[node.Id] = *node
			delete(expired, node.Id)
		} else {
			_, ok := new[node.Id]
			if !ok {
				current[node.Id] = *node
				new[node.Id] = node
				updated = true
			}
		}
	}

	if len(expired) != 0 {
		updated = true
	}

	return current, updated
}

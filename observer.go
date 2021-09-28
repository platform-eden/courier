package courier

import (
	"log"
	"time"
)

type storeObserver struct {
	store           NodeStorer
	observeInterval time.Duration
	currentNodes    map[string]*Node
	nodeChannels    []chan (map[string]*Node)
	subjects        []string
}

// Returns a paused StoreObserver
func NewStoreObserver(store NodeStorer, interval time.Duration, subjects []string) *storeObserver {
	s := storeObserver{
		store:           store,
		observeInterval: interval,
		currentNodes:    map[string]*Node{},
		nodeChannels:    []chan (map[string]*Node){},
		subjects:        subjects,
	}

	return &s
}

// Adds a channel to the StoreObserver that will receive a map of Nodes when the NodeStore has updated Nodes and returns it
func (s *storeObserver) listenChannel() chan (map[string]*Node) {
	channel := make(chan map[string]*Node)
	s.nodeChannels = append(s.nodeChannels, channel)

	return channel
}

// Starts a Goroutine that will begin comparing current nodes and what nodes the NodeStore has.  If the NodeStore updates,
// it sends a new map of Nodes to each Node Channel listening to the Observer.
func (s *storeObserver) start() {
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
				for _, channel := range s.nodeChannels {
					channel <- current
				}
			}
		}
	}()
}

// compares the Nodes returned from the NodeStore with the current Nodes in the service.
// If there are differences, this will return true with an updated map of Nodes.
func compareNodes(potential []*Node, expired map[string]*Node) (map[string]*Node, bool) {
	current := map[string]*Node{}
	new := map[string]*Node{}
	updated := false

	for _, node := range potential {
		_, ok := expired[node.Id]

		if ok {
			current[node.Id] = node
			delete(expired, node.Id)
		} else {
			_, ok := new[node.Id]
			if !ok {
				current[node.Id] = node
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

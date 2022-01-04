package messaging

import (
	"context"
	"log"
	"sync"
)

type Noder interface {
	Id() string
	Address() string
	Subscribed() []string
	Broadcasted() []string
	Port() string
}

type NodeMapper interface {
	Node(string) (Node, bool)
	Nodes() map[string]Node
	Update(...Node)
	Add(Node)
	Remove(string)
	Length() int
}
type nodeRegistry struct {
	observeChannel <-chan []Noder
	newChannel     chan Node
	failedChannel  chan Node
	staleChannel   chan Node
	blacklist      NodeMapper
	current        NodeMapper
	waitGroup      *sync.WaitGroup
	cancelFunc     context.CancelFunc
}

type nodeRegistryOptions struct {
	observeChannel <-chan []Noder
	newChannel     chan Node
	failedChannel  chan Node
	staleChannel   chan Node
	startRegistry  bool
}

func newNodeRegistry(options *nodeRegistryOptions) *nodeRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	r := &nodeRegistry{
		observeChannel: options.observeChannel,
		newChannel:     options.newChannel,
		failedChannel:  options.failedChannel,
		staleChannel:   options.staleChannel,
		waitGroup:      wg,
		cancelFunc:     cancel,
		blacklist:      NewNodeMap(),
		current:        NewNodeMap(),
	}

	if options.startRegistry {
		go r.registerNodes(ctx, wg)
	}
	return r
}

func (registry *nodeRegistry) stop() {
	registry.cancelFunc()
	registry.waitGroup.Wait()

	// close the channels we write on
	close(registry.newChannel)
	close(registry.staleChannel)
}

// registerNodes either receives new nodes to be sifted and sent out of the newChannel or receives nodes that could not receive a message that need to be blacklisted.
func (registry *nodeRegistry) registerNodes(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case noders := <-registry.observeChannel:
			blacklisted, cll := updateNodes(noders, registry.newChannel, registry.staleChannel, registry.blacklist, registry.current)
			registry.blacklist.Update(blacklisted...)
			registry.current.Update(cll...)

		case n := <-registry.failedChannel:
			registry.blacklist.Add(n)
			registry.current.Remove(n.id)
		}
	}
}

// updateNodes compares a list of nodes with blacklisted nodes and current nodes.  Any new nodes are sent through the nodeChannel.  Returns a new blacklist and current node list.
func updateNodes(noderList []Noder, nchan chan Node, schan chan Node, blacklist NodeMapper, current NodeMapper) ([]Node, []Node) {
	blacklisted := make(chan Node, blacklist.Length()+1)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	noders := generateNoders(noderList...)
	nodes := nodersToNodes(noders)
	cblout := compareBlackList(nodes, blacklisted, blacklist)
	new, active, stale := compareCurrentList(cblout, current)

	buffactive := nodeBuffer(active, len(noderList)+1)
	go sendNodes(new, nchan, wg)
	go sendNodes(stale, schan, wg)

	wg.Wait()

	return nodeChannelToList(blacklisted), nodeChannelToList(buffactive)
}

func nodeChannelToList(in <-chan Node) []Node {
	nodeList := []Node{}
	for n := range in {
		nodeList = append(nodeList, n)
	}

	return nodeList
}

func generateNoders(noders ...Noder) <-chan Noder {
	out := make(chan Noder)
	go func() {
		for _, n := range noders {
			out <- n
		}
		close(out)
	}()

	return out
}

func nodersToNodes(in <-chan Noder) <-chan Node {
	out := make(chan Node)
	go func() {
		for n := range in {
			out <- *NewNode(n.Id(), n.Address(), n.Port(), n.Subscribed(), n.Broadcasted())
		}
		close(out)
	}()

	return out
}

// compareBlackList compares incoming nodes to a map of nodes.  If the node is in the map, this function logs to output.
// If a node does not exist in the blacklist, it is passed through the returned node channel.
func compareBlackList(in <-chan Node, blacklisted chan Node, blacklist NodeMapper) <-chan Node {
	out := make(chan Node)
	go func() {
		for n := range in {
			if blacklist.Length() != 0 {
				_, exist := blacklist.Node(n.id)
				if exist {
					log.Printf("Node %s is currently blacklisted - skipping node", n.id)
					blacklisted <- n
					continue
				}
			}

			out <- n
		}

		close(out)
		close(blacklisted)
	}()

	return out
}

// compareCurrentList compares nodes coming in to a map of nodes. All nodes passed in are passed out through the
// active channel and all new nodes are passed through the returned node channel.
func compareCurrentList(in <-chan Node, current NodeMapper) (<-chan Node, <-chan Node, <-chan Node) {
	out := make(chan Node)
	active := make(chan Node)
	stale := make(chan Node)
	go func() {
		for n := range in {
			active <- n
			_, exist := current.Node(n.id)
			if exist {
				current.Remove(n.id)
				continue
			}
			out <- n
		}

		close(out)
		close(active)

		for _, v := range current.Nodes() {
			stale <- v
		}

		close(stale)
	}()

	return out, active, stale
}

// sendNodes sends nodes passed to it out to the nodeChannel. Returns a channel that will return true once
// the in channel is closed and sendNodes is complete
func sendNodes(in <-chan Node, nodeChannel chan Node, wg *sync.WaitGroup) {
	for n := range in {
		nodeChannel <- n
	}

	defer wg.Done()
}

// nodeBuffer takes a regular channel and returns a buffered channel
func nodeBuffer(in <-chan Node, size int) <-chan Node {
	out := make(chan Node, size)

	go func() {
		for n := range in {
			out <- n
		}
		close(out)
	}()

	return out
}

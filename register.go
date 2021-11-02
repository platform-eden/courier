package courier

import (
	"log"
	"sync"

	"github.com/platform-edn/courier/node"
)

// registerNodes either receives new nodes to be sifted and sent out of the newChannel or receives nodes that could not receive a message that need to be blacklisted.
func registerNodes(ochan <-chan []node.Node, nchan chan node.Node, schan chan node.Node, fchan <-chan node.Node, blacklist NodeMapper, current NodeMapper) {
	for {
		select {
		case nodes := <-ochan:
			blacklisted, cll := updateNodes(nodes, nchan, schan, blacklist, current)
			blacklist.Update(<-blacklisted)
			current.Update(<-cll)

		case n := <-fchan:
			blacklist.Add(n)
			current.Remove(n.Id)
		}
	}

	//need to implement of graceful shutdown
	//close(newChannel)
	//close(staleChannel)
}

// updateNodes compares a list of nodes with blacklisted nodes and current nodes.  Any new nodes are sent through the nodeChannel.  Returns a new blacklist and current node list.
func updateNodes(nodes []node.Node, nchan chan node.Node, schan chan node.Node, blacklist NodeMapper, current NodeMapper) (<-chan node.Node, <-chan node.Node) {
	blacklisted := make(chan node.Node, blacklist.Length()+1)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	gout := generate(nodes...)
	cblout := compareBlackList(gout, blacklisted, blacklist)
	new, active, stale := compareCurrentList(cblout, current)

	buffactive := nodeBuffer(active, len(nodes)+1)
	go sendNodes(new, nchan, wg)
	go sendNodes(stale, schan, wg)

	wg.Wait()

	return blacklisted, buffactive
}

// generate takes a set of nodes and returns a channel of those nodes
func generate(nodes ...node.Node) <-chan node.Node {
	out := make(chan node.Node)
	go func() {
		for _, n := range nodes {
			out <- n
		}
		close(out)
	}()

	return out
}

// compareBlackList compares incoming nodes to a map of nodes.  If the node is in the map, this function logs to output.
// If a node does not exist in the blacklist, it is passed through the returned node channel.
func compareBlackList(in <-chan node.Node, blacklisted chan node.Node, blacklist NodeMapper) <-chan node.Node {
	out := make(chan node.Node)
	go func() {
		for n := range in {
			if blacklist.Length() != 0 {
				_, exist := blacklist.Node(n.Id)
				if exist {
					log.Printf("Node %s is currently blacklisted - skipping node", n.Id)
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
func compareCurrentList(in <-chan node.Node, current NodeMapper) (<-chan node.Node, <-chan node.Node, <-chan node.Node) {
	out := make(chan node.Node)
	active := make(chan node.Node)
	stale := make(chan node.Node)
	go func() {
		for n := range in {
			active <- n
			_, exist := current.Node(n.Id)
			if exist {
				current.Remove(n.Id)
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
func sendNodes(in <-chan node.Node, nodeChannel chan node.Node, wg *sync.WaitGroup) {
	for n := range in {
		nodeChannel <- n
	}

	defer wg.Done()
}

// nodeBuffer takes a regular channel and returns a buffered channel
func nodeBuffer(in <-chan node.Node, size int) <-chan node.Node {
	out := make(chan node.Node, size)

	go func() {
		for n := range in {
			out <- n
		}
		close(out)
	}()

	return out
}

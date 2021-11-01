package courier

import (
	"log"

	"github.com/platform-edn/courier/node"
)

func registerNodes(observeChannel <-chan []node.Node, newChannel chan node.Node, failedConnectionChannel <-chan node.Node, blacklist NodeMapper, current NodeMapper) {
	for {
		select {
		case nodes := <-observeChannel:
			bll, cll := updateNodes(nodes, newChannel, blacklist, current)
			blacklist.Update(bll...)
			current.Update(cll...)

		case n := <-failedConnectionChannel:
			blacklist.Add(n)
			current.Remove(n.Id)
		}
	}

	//need to implement of graceful shutdown
	//close(newChannel)
}

func updateNodes(nodes []node.Node, nodeChannel chan node.Node, blacklist NodeMapper, current NodeMapper) ([]node.Node, []node.Node) {
	blacklisted := make(chan node.Node, blacklist.Length()+1)
	active := make(chan node.Node, len(nodes)+1)

	gout := generate(nodes...)
	cblout := compareBlackList(gout, blacklisted, blacklist)
	cclout := compareCurrentList(cblout, active, current)
	done := sendNodes(cclout, nodeChannel)

	<-done

	bll := []node.Node{}
	for n := range blacklisted {
		bll = append(bll, n)
	}

	cll := []node.Node{}
	for n := range active {
		cll = append(cll, n)
	}

	return bll, cll
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

// compareBlackList comapres incoming nodes to a map of nodes.  If the node is in the map, this function logs to output.
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
func compareCurrentList(in <-chan node.Node, active chan node.Node, current NodeMapper) <-chan node.Node {
	out := make(chan node.Node)
	go func() {
		for n := range in {
			active <- n
			_, exist := current.Node(n.Id)
			if exist {
				continue
			}
			out <- n
		}
		close(out)
		close(active)
	}()

	return out
}

// sendNodes sends nodes passed to it out to the newNodes channel. Returns a channel that will return true once
// the in channel is closed and sendNodes is complete
func sendNodes(in <-chan node.Node, newNodes chan node.Node) <-chan bool {
	done := make(chan bool)
	go func() {
		for n := range in {
			newNodes <- n
		}

		done <- true
		close(done)
	}()

	return done
}

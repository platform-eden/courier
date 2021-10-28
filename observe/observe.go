package observe

import (
	"log"
	"time"

	"github.com/platform-edn/courier/node"
)

// Starts a Goroutine that will begin comparing current nodes and what nodes the NodeStore has.  If the NodeStore updates,
// it sends a new map of Nodes to each Node Channel listening to the Observer.
func Observe(options ObserveOptions) {
	blacklist := NewNodeMap()
	current := NewNodeMap()
	blacklisted := make(chan node.Node)
	defer close(blacklisted)
	active := make(chan node.Node)
	defer close(active)

	for {
		select {
		case <-time.After(options.Interval):
			nodes, err := options.Store.GetSubscribers(options.Subjects...)
			if err != nil {
				log.Printf("could not observe Courier nodes: %s", err)
				return
			}

			gout := generate(nodes...)
			cblout := compareBlackList(gout, blacklisted, blacklist)
			cclout := compareCurrentList(cblout, active, current)
			sent := sendNodes(cclout, options.NewNodeChannel)

			tblacklist := NewNodeMap()
			tcurrent := NewNodeMap()
			done := false

			for !done {
				select {
				case b := <-blacklisted:
					tblacklist.Add(b)
				case a := <-active:
					tcurrent.Add(a)
				case <-sent:
					done = true
				}
			}

			blacklist = tblacklist
			current = tcurrent

		case n := <-options.FailedConnectionChannel:
			blacklist.Add(n)
			current.Remove(n.Id)
		}
	}
}

func generate(nodes ...*node.Node) <-chan node.Node {
	out := make(chan node.Node)
	go func() {
		for _, n := range nodes {
			out <- *n
		}
		close(out)
	}()

	return out
}

func compareBlackList(in <-chan node.Node, blacklisted chan node.Node, blacklist node.NodeMapper) <-chan node.Node {
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
	}()

	return out
}

func compareCurrentList(in <-chan node.Node, active chan node.Node, current node.NodeMapper) <-chan node.Node {
	out := make(chan node.Node)
	go func() {
		for n := range in {
			_, exist := current.Node(n.Id)
			if exist {
				continue
			}
			out <- n
		}
		close(out)
	}()

	return out
}

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

package observer

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/mock"
	"github.com/platform-edn/courier/node"
)

func TestStoreObserver_Observe(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := mock.CreateTestNodes(nodecount, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})
	store := mock.NewMockNodeStore(nodes...)
	observer := NewStoreObserver(store, (time.Second * 1), subjects)

	nodeChannel := observer.NodeChannel()

	timer := time.NewTimer(time.Second * 3)

	select {
	case <-timer.C:
		t.Fatalf("observer never added new nodes")
	case <-nodeChannel:
		timer.Stop()
	}

	newNode := mock.CreateTestNodes(1, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]
	store.AddNode(newNode)

	timer.Reset(time.Second * 3)

	select {
	case <-timer.C:
		t.Fatalf("observer never added new nodes")
	case <-nodeChannel:
		timer.Stop()
	}

	observer.failedConnections <- *nodes[0]

	timer.Reset(time.Second * 3)

blacklistloop:
	for {
		select {
		case <-timer.C:
			t.Fatalf("observer never added new nodes")
		default:
			if observer.blackListedNodes.Length() > 0 && observer.currentNodes.Length() == nodecount-1 {
				break blacklistloop
			}
		}
	}
}

func TestCompareBlackListNodes(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := mock.CreateTestNodes(nodecount, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})

	blacklist := map[string]node.Node{}
	skip := true

	for _, n := range nodes {
		if skip == true {
			skip = false
			continue
		}
		skip = true
		blacklist[n.Id] = *n
	}

	newNodes, newBlacklist := compareBlackListNodes(nodes, blacklist)
	if len(newNodes) != nodecount/2 {
		t.Fatalf("expected length of returned nodes to be %v but got %v", nodecount/2, len(newNodes))
	}
	if len(newBlacklist) != nodecount/2 {
		t.Fatalf("expected length of returned blacklist to be %v but got %v", nodecount/2, len(newBlacklist))
	}

	for _, n := range newNodes {
		for _, b := range newBlacklist {
			if b.Id == n.Id {
				t.Fatalf("Node with id %s is in both returned nodes and returned blacklist", n.Id)
			}
		}
	}

	_, newBlacklist = compareBlackListNodes(newNodes, newBlacklist)
	if len(newBlacklist) != 0 {
		t.Fatalf("expected length of returned blacklist to be %v but got %v", 0, len(newBlacklist))

	}
}

func TestComparePotentialNodes(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := mock.CreateTestNodes(nodecount, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})

	nodeMap := map[string]node.Node{}
	for _, n := range nodes {
		nodeMap[n.Id] = *n
	}

	removedNode := nodes[0]
	addedNode := mock.CreateTestNodes(1, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]
	nodes[0] = addedNode

	current, updated := comparePotentialNodes(nodes, nodeMap)
	if !updated {
		t.Fatalf("should have returned updated as true but got false")
	}

	for _, n := range current {
		if n.Id == removedNode.Id {
			t.Fatalf("expected node %s to be removed but it wasn't", removedNode.Id)

		}
	}

	_, updated = comparePotentialNodes(nodes, current)
	if updated {
		t.Fatalf("expected updated to be false but got true")
	}

}

func TestSendUpdateNodes(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := mock.CreateTestNodes(nodecount, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})
	nodeMap := map[string]node.Node{}
	nchan := make(chan map[string]node.Node)
	nchan1 := make(chan map[string]node.Node)
	nchans := []chan map[string]node.Node{nchan, nchan1}

	for _, n := range nodes {
		nodeMap[n.Id] = *n
	}

	go sendUpdatedNodes(nodeMap, nchans)

	nt := false
	nt1 := false

listenloop:
	for {
		select {
		case <-nchan:
			nt = true
		case <-nchan1:
			nt1 = true
		case <-time.After(time.Second * 1):
			t.Fatal("timeout waitng on node channel")
		default:
			if nt && nt1 {
				break listenloop
			}
		}
	}
}

package observer

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/node"
	test "github.com/platform-edn/courier/testing"
)

func TestStoreObserver_Start(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := test.CreateTestNodes(nodecount, &test.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})
	store := test.NewMockNodeStore(nodes...)
	observer := NewStoreObserver(store, (time.Second * 1), subjects)

	nodeChannel := observer.ListenChannel()

	observer.Start()

	timer := time.NewTimer(time.Second * 3)

	select {
	case <-timer.C:
		t.Fatalf("observer never added new nodes")
	case <-nodeChannel:
		timer.Stop()
	}

	newNode := test.CreateTestNodes(1, &test.TestNodeOptions{
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
}

func TestCompareNodes(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := test.CreateTestNodes(nodecount, &test.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})

	nodeMap := map[string]*node.Node{}

	for _, n := range nodes {
		nodeMap[n.Id] = n
	}

	removedNode := nodes[0]
	addedNode := test.CreateTestNodes(1, &test.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]
	nodes[0] = addedNode

	current, updated := compareNodes(nodes, nodeMap)
	if !updated {
		t.Fatalf("should have returned updated as true but got false")
	}

	for _, n := range current {
		if n.Id == removedNode.Id {
			t.Fatalf("expected node %s to be removed but it wasn't", removedNode.Id)

		}
	}

	_, updated = compareNodes(nodes, current)
	if updated {
		t.Fatalf("expected updated to be false but got true")
	}
}

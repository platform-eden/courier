package courier

import (
	"testing"
	"time"
)

func TestStoreObserver_Start(t *testing.T) {
	nodecount := 10
	subjects := []string{"sub1", "sub2", "sub3"}
	nodes := createTestNodes(nodecount, &testNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})
	store := NewMockNodeStore(nodes...)
	observer := NewStoreObserver(store, (time.Second * 1), subjects)

	nodeChannel := observer.listenChannel()

	observer.start()

	timer := time.NewTimer(time.Second * 3)

	select {
	case <-timer.C:
		t.Fatalf("observer never added new nodes")
	case <-nodeChannel:
		timer.Stop()
	}

	newNode := createTestNodes(1, &testNodeOptions{
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
	nodes := createTestNodes(nodecount, &testNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})

	nodeMap := map[string]*Node{}

	for _, node := range nodes {
		nodeMap[node.Id] = node
	}

	removedNode := nodes[0]
	addedNode := createTestNodes(1, &testNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]
	nodes[0] = addedNode

	current, updated := compareNodes(nodes, nodeMap)
	if !updated {
		t.Fatalf("should have returned updated as true but got false")
	}

	for _, node := range current {
		if node.Id == removedNode.Id {
			t.Fatalf("expected node %s to be removed but it wasn't", removedNode.Id)

		}
	}

	_, updated = compareNodes(nodes, current)
	if updated {
		t.Fatalf("expected updated to be false but got true")
	}
}

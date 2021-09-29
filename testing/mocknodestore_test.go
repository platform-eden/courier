package testing

import (
	"testing"

	"github.com/platform-edn/courier/node"
)

func TestNewMockNodeStore(t *testing.T) {
	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	nodes := CreateTestNodes(5, &TestNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})

	store := NewMockNodeStore(nodes...)

	for _, n := range nodes {

		for _, subject := range n.SubscribedSubjects {
			subnodes, ok := store.SubscribeNodes[subject]
			if !ok {
				t.Fatalf("expected to find node %s subscribed to subject but subject %s doesn't exist", n.Id, subject)
			}

			found := false
			for _, s := range subnodes {
				if s.Id == n.Id {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", n.Id, subject)
			}
		}

		for _, subject := range n.BroadcastedSubjects {
			broadnodes, ok := store.BroadCastNodes[subject]
			if !ok {
				t.Fatalf("expected to find node %s broadcasting to subject %s but it was not found", n.Id, subject)
			}
			found := false
			for _, s := range broadnodes {
				if s.Id == n.Id {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", n.Id, subject)
			}
		}
	}
}

func TestMockNodeStore_GetSubscribers(t *testing.T) {
	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	nodes := CreateTestNodes(10, &TestNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})

	subnodes := []*node.Node{}

	for _, n := range nodes {
		for _, subject := range n.SubscribedSubjects {
			if subject == "sub2" {
				subnodes = append(subnodes, n)
			}

			if subject == "sub3" {
				subnodes = append(subnodes, n)
			}
		}
	}

	store := NewMockNodeStore(nodes...)

	storenodes, _ := store.GetSubscribers("sub2", "sub3")

	if len(storenodes) != len(subnodes) {
		t.Fatalf("expected length of nodes returned to be %v but got %v", len(subnodes), len(storenodes))
	}
}

func TestMockNodeStore_AddNode(t *testing.T) {
	store := NewMockNodeStore()

	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	n := CreateTestNodes(1, &TestNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})[0]

	store.AddNode(n)

	for _, subject := range n.SubscribedSubjects {
		_, ok := store.SubscribeNodes[subject]
		if !ok {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", n.Id, subject)
		}
	}

	for _, subject := range n.BroadcastedSubjects {
		_, ok := store.BroadCastNodes[subject]
		if !ok {
			t.Fatalf("expected to find node %s broadcasting to subject %s but it was not found", n.Id, subject)
		}
	}

	store.AddNode(n)
}

func TestMockNodeStore_RemoveNode(t *testing.T) {
	store := NewMockNodeStore()

	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	n := CreateTestNodes(1, &TestNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})[0]

	store.RemoveNode(n)
	store.AddNode(n)
	store.RemoveNode(n)

	for _, subject := range n.SubscribedSubjects {
		subnodes := store.SubscribeNodes[subject]
		if len(subnodes) != 0 {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", n.Id, subject)
		}
	}

	for _, subject := range n.BroadcastedSubjects {
		broadnodes := store.BroadCastNodes[subject]
		if len(broadnodes) != 0 {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", n.Id, subject)
		}
	}

	store.RemoveNode(n)
}

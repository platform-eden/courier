package courier

import (
	"testing"
)

func TestNewMockNodeStore(t *testing.T) {
	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	nodes := createTestNodes(5, &testNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})

	store := NewMockNodeStore(nodes...)

	for _, node := range nodes {

		for _, subject := range node.SubscribedSubjects {
			subnodes, ok := store.SubscribeNodes[subject]
			if !ok {
				t.Fatalf("expected to find node %s subscribed to subject but subject %s doesn't exist", node.Id, subject)
			}

			found := false
			for _, s := range subnodes {
				if s.Id == node.Id {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", node.Id, subject)
			}
		}

		for _, subject := range node.BroadcastedSubjects {
			broadnodes, ok := store.BroadCastNodes[subject]
			if !ok {
				t.Fatalf("expected to find node %s broadcasting to subject %s but it was not found", node.Id, subject)
			}
			found := false
			for _, s := range broadnodes {
				if s.Id == node.Id {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", node.Id, subject)
			}
		}
	}
}

func TestMockNodeStore_GetSubscribers(t *testing.T) {
	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	nodes := createTestNodes(10, &testNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})

	subnodes := []*Node{}

	for _, node := range nodes {
		for _, subject := range node.SubscribedSubjects {
			if subject == "sub2" {
				subnodes = append(subnodes, node)
			}

			if subject == "sub3" {
				subnodes = append(subnodes, node)
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

	node := createTestNodes(1, &testNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})[0]

	store.AddNode(node)

	for _, subject := range node.SubscribedSubjects {
		_, ok := store.SubscribeNodes[subject]
		if !ok {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", node.Id, subject)
		}
	}

	for _, subject := range node.BroadcastedSubjects {
		_, ok := store.BroadCastNodes[subject]
		if !ok {
			t.Fatalf("expected to find node %s broadcasting to subject %s but it was not found", node.Id, subject)
		}
	}

	store.AddNode(node)
}

func TestMockNodeStore_RemoveNode(t *testing.T) {
	store := NewMockNodeStore()

	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	node := createTestNodes(1, &testNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})[0]

	store.RemoveNode(node)
	store.AddNode(node)
	store.RemoveNode(node)

	for _, subject := range node.SubscribedSubjects {
		subnodes := store.SubscribeNodes[subject]
		if len(subnodes) != 0 {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", node.Id, subject)
		}
	}

	for _, subject := range node.BroadcastedSubjects {
		broadnodes := store.BroadCastNodes[subject]
		if len(broadnodes) != 0 {
			t.Fatalf("expected to find node %s subscribed to subject %s but it was not found", node.Id, subject)
		}
	}

	store.RemoveNode(node)
}

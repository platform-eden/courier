package client

import (
	"math/rand"
	"testing"

	"github.com/platform-edn/courier/mock"
	"github.com/platform-edn/courier/node"
)

func TestSubscriberMap_AddSubscriber(t *testing.T) {
	subMap := NewSubscriberMap()
	subjects := []string{"sub", "sub1", "sub2"}

	nodes := mock.CreateTestNodes(10, &mock.TestNodeOptions{
		SubscribedSubjects: subjects,
	})
	node := nodes[0]

	subMap.AddSubscriber(node)

	for _, subject := range node.SubscribedSubjects {
		if len(subMap.subscribers[subject]) == 0 {
			t.Fatalf("expected length of map to be %v but got %v", 1, len(subMap.subscribers))
		}
	}

	subMap.AddSubscriber(node)

	for _, subject := range node.SubscribedSubjects {
		if len(subMap.subscribers[subject]) != 1 {
			t.Fatalf("expected length of map to be %v but got %v", 1, len(subMap.subscribers))
		}
	}
}

func TestSubscriberMap_RemoveSubscriber(t *testing.T) {
	subMap := NewSubscriberMap()
	subjects := []string{"sub", "sub1", "sub2"}
	nodes := mock.CreateTestNodes(2, &mock.TestNodeOptions{
		SubscribedSubjects: subjects,
	})
	node1 := nodes[0]
	node2 := nodes[1]

	node1.SubscribedSubjects = subjects
	node2.SubscribedSubjects = subjects

	for _, subject := range subjects {
		subMap.subscribers[subject] = append(subMap.subscribers[subject], node1)
		subMap.subscribers[subject] = append(subMap.subscribers[subject], node2)
	}

	err := subMap.RemoveSubscriber(node1)
	if err != nil {
		t.Fatalf("expected remove subscriber to pass but it failed: %s", err)
	}

	for _, subject := range subjects {
		sublist, ok := subMap.subscribers[subject]
		if !ok {
			t.Fatalf("expected subject %s to exist but it doesn't", subject)
		}

		if len(sublist) != 1 {
			t.Fatalf("expect subject list to be 1 but got %v", len(sublist))
		}
	}

	err = subMap.RemoveSubscriber(node2)
	if err != nil {
		t.Fatalf("expected remove subscriber to pass but it failed: %s", err)
	}

	for _, subject := range subjects {
		_, ok := subMap.subscribers[subject]
		if ok {
			t.Fatalf("expect subject %s to not exist but it does", subject)
		}
	}
}

func TestSubscriberMap_SubjectSubscribers(t *testing.T) {
	length := 5
	subjects := []string{"sub", "sub1"}
	nodes := mock.CreateTestNodes(length, &mock.TestNodeOptions{
		SubscribedSubjects: subjects,
	})
	subMap := NewSubscriberMap()
	subnodes := []*node.Node{}
	sub1nodes := []*node.Node{}

	for _, node := range nodes {
		subMap.AddSubscriber(node)
		for _, subject := range node.SubscribedSubjects {
			if subject == "sub" {
				subnodes = append(subnodes, node)
			} else {
				sub1nodes = append(sub1nodes, node)
			}
		}
	}

	subscribers, err := subMap.SubjectSubscribers("sub")
	if err != nil {
		t.Fatalf("expected getting subscribers to pass but it failed: %s", err)
	}
	subscribers1, err := subMap.SubjectSubscribers("sub1")
	if err != nil {
		t.Fatalf("expected getting subscribers to pass but it failed: %s", err)
	}

	_, err = subMap.SubjectSubscribers("sub2")
	if err == nil {
		t.Fatal("expected getting subscribers to fail when fetching nonexistant subject but it passed")
	}

	if len(subscribers) != len(subnodes) {
		t.Fatalf("expected list of subscribers to have a length of %v but got %v", len(subnodes), len(subscribers))
	}
	if len(subscribers1) != len(sub1nodes) {
		t.Fatalf("expected list of subscribers to have a length of %v but got %v", len(sub1nodes), len(subscribers1))
	}
}

func TestSubscriberMap_GetAllSubscribers(t *testing.T) {
	length := 5
	nodes := mock.CreateTestNodes(length, &mock.TestNodeOptions{})
	subMap := NewSubscriberMap()

	for _, node := range nodes {
		subMap.AddSubscriber(node)
	}

	subscribers := subMap.AllSubscribers()

	if len(subscribers) != length {
		t.Fatalf("expected list of subscribers to have a length of %v but got %v", length, len(nodes))
	}
}

func TestRemoveSubscriberFromSubject(t *testing.T) {
	nodes := mock.CreateTestNodes(5, &mock.TestNodeOptions{})
	np := rand.Intn(5)
	node := nodes[np]
	length := len(nodes)

	_, err := removeSubscriberFromSubject(nodes, -1)
	if err == nil {
		t.Fatal("expected removeSubscriberFromSubject to fail from negative but it passed")
	}

	_, err = removeSubscriberFromSubject(nodes, 20)
	if err == nil {
		t.Fatal("expected removeSubscriberFromSubject to fail from high number but it passed")
	}

	nodes, err = removeSubscriberFromSubject(nodes, np)
	if err != nil {
		t.Fatalf("expected removeSubscriberFromSubjecto pass but it failed: %s", err)
	}
	if length <= len(nodes) {
		t.Fatalf("expected length of nodes to be %v but got %v", length-1, len(nodes))
	}

	_, ok := findSubscriber(nodes, node.Id)
	if ok {
		t.Fatal("expected node to be removed but it was not")

	}
}

func TestFindSubscriber(t *testing.T) {
	nodes := mock.CreateTestNodes(5, &mock.TestNodeOptions{})
	np := rand.Intn(5)
	node := nodes[np]

	position, exists := findSubscriber(nodes, node.Id)
	if !exists {
		t.Fatal("expected node to exist but it doesn't")
	}
	if position != np {
		t.Fatalf("expected node to be found at %v but was found at %v", np, position)
	}

	node = mock.CreateTestNodes(1, &mock.TestNodeOptions{})[0]

	_, exists = findSubscriber(nodes, node.Id)
	if exists {
		t.Fatal("expected node to not exist but it does")
	}

	/*TODO*/
	//should check to make sure that we're getting the correct amount of nodes in each subject
}

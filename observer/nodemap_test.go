package observer

import (
	"testing"

	"github.com/platform-edn/courier/mock"
	"github.com/platform-edn/courier/node"
)

func TestNodeMap_Nodes(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mock.CreateTestNodes(l, &mock.TestNodeOptions{})

	for _, n := range nodes {
		b.AddNode(*n)
	}

	for _, n := range nodes {
		_, exist := b.nodes[n.Id]
		if !exist {
			t.Fatalf("expected node with id %s to exist but it didn't", n.Id)
		}
	}
}

func TestNodeMap_Update(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mock.CreateTestNodes(l, &mock.TestNodeOptions{})

	nmap := map[string]node.Node{}

	for _, n := range nodes {
		nmap[n.Id] = *n
	}

	b.Update(nmap)

	for _, n := range nodes {
		_, exist := b.nodes[n.Id]
		if !exist {
			t.Fatalf("expected node with id %s to exist but it didn't", n.Id)
		}
	}
}

func TestNodeMap_AddNode(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mock.CreateTestNodes(l, &mock.TestNodeOptions{})

	for _, n := range nodes {
		b.AddNode(*n)
	}

	bl := b.Length()

	if bl != l {
		t.Fatalf("expected length to be %v but got %v", l, bl)
	}
}

func TestNodeMap_RemoveNode(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mock.CreateTestNodes(l, &mock.TestNodeOptions{})

	for _, n := range nodes {
		b.AddNode(*n)
	}

	for _, n := range nodes {
		b.RemoveNode(n.Id)
	}

	bl := b.Length()

	if bl != 0 {
		t.Fatalf("expected length to be %v but got %v", 0, bl)
	}
}

func TestNodeMap_Length(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mock.CreateTestNodes(l, &mock.TestNodeOptions{})

	for _, n := range nodes {
		b.AddNode(*n)
	}

	bl := b.Length()

	if bl != len(b.nodes) {
		t.Fatalf("expected length to be %v but got %v", len(b.nodes), bl)
	}
}

package registry

import (
	"testing"
)

func TestNodeMap_Nodes(t *testing.T) {
	l := 10
	nodes := CreateTestNodes(l, &TestNodeOptions{})
	b := NewNodeMap(RemovePointers(nodes)...)

	for _, n := range nodes {
		_, exist := b.nodes[n.Id]
		if !exist {
			t.Fatalf("expected node with id %s to exist but it didn't", n.Id)
		}
	}
}

func TestNodeMap_Add(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := CreateTestNodes(l, &TestNodeOptions{})

	for _, n := range nodes {
		b.addNode(*n)
	}

	bl := b.Length()

	if bl != l {
		t.Fatalf("expected length to be %v but got %v", l, bl)
	}
}

func TestNodeMap_Remove(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := CreateTestNodes(l, &TestNodeOptions{})

	for _, n := range nodes {
		b.addNode(*n)
	}

	for _, n := range nodes {
		b.removeNode(n.Id)
	}

	bl := b.Length()

	if bl != 0 {
		t.Fatalf("expected length to be %v but got %v", 0, bl)
	}
}

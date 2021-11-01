package courier

import (
	"testing"

	"github.com/platform-edn/courier/mocks"
)

func TestNodeMap_Nodes(t *testing.T) {
	l := 10
	nodes := mocks.CreateTestNodes(l, &mocks.TestNodeOptions{})
	b := NewNodeMap(mocks.RemovePointers(nodes)...)

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
	nodes := mocks.CreateTestNodes(l, &mocks.TestNodeOptions{})

	for _, n := range nodes {
		b.Add(*n)
	}

	bl := b.Length()

	if bl != l {
		t.Fatalf("expected length to be %v but got %v", l, bl)
	}
}

func TestNodeMap_Remove(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mocks.CreateTestNodes(l, &mocks.TestNodeOptions{})

	for _, n := range nodes {
		b.Add(*n)
	}

	for _, n := range nodes {
		b.Remove(n.Id)
	}

	bl := b.Length()

	if bl != 0 {
		t.Fatalf("expected length to be %v but got %v", 0, bl)
	}
}

func TestNodeMap_Length(t *testing.T) {
	b := NewNodeMap()
	l := 10
	nodes := mocks.CreateTestNodes(l, &mocks.TestNodeOptions{})

	for _, n := range nodes {
		b.Add(*n)
	}

	bl := b.Length()

	if bl != len(b.nodes) {
		t.Fatalf("expected length to be %v but got %v", len(b.nodes), bl)
	}
}

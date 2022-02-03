package client

import (
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
)

func TestClientNodeMap_Nodes(t *testing.T) {
	l := 10
	nodes := registry.CreateTestNodes(l, &registry.TestNodeOptions{})
	b := newClientNodeMap()

	for _, n := range nodes {
		cn, err := newClientNode(*n, uuid.NewString(), []ClientNodeOption{WithDialOptions(grpc.WithInsecure())}...)
		if err != nil {
			t.Fatalf("could not create client node: %s", err)
		}
		b.AddClientNode(*cn)
	}

	for _, n := range nodes {
		_, exist := b.nodes[n.Id]
		if !exist {
			t.Fatalf("expected node with id %s to exist but it didn't", n.Id)
		}
	}
}

func TestClientNodeMap_Add(t *testing.T) {
	b := newClientNodeMap()
	l := 10
	nodes := registry.CreateTestNodes(l, &registry.TestNodeOptions{})

	for _, n := range nodes {
		cn, err := newClientNode(*n, uuid.NewString(), []ClientNodeOption{WithDialOptions(grpc.WithInsecure())}...)
		if err != nil {
			t.Fatalf("could not create client node: %s", err)
		}
		b.AddClientNode(*cn)
	}

	bl := b.Length()

	if bl != l {
		t.Fatalf("expected length to be %v but got %v", l, bl)
	}
}

func TestClientNodeMap_Remove(t *testing.T) {
	b := newClientNodeMap()
	l := 10
	nodes := registry.CreateTestNodes(l, &registry.TestNodeOptions{})

	for _, n := range nodes {
		cn, err := newClientNode(*n, uuid.NewString(), []ClientNodeOption{WithDialOptions(grpc.WithInsecure())}...)
		if err != nil {
			t.Fatalf("could not create client node: %s", err)
		}
		b.AddClientNode(*cn)
	}

	for _, n := range nodes {
		b.RemoveClientNode(n.Id)
	}

	bl := b.Length()

	if bl != 0 {
		t.Fatalf("expected length to be %v but got %v", 0, bl)
	}
}

func TestClientNodeMap_Length(t *testing.T) {
	b := newClientNodeMap()
	l := 10
	nodes := registry.CreateTestNodes(l, &registry.TestNodeOptions{})

	for _, n := range nodes {
		cn, err := newClientNode(*n, uuid.NewString(), []ClientNodeOption{WithDialOptions(grpc.WithInsecure())}...)
		if err != nil {
			t.Fatalf("could not create client node: %s", err)
		}
		b.AddClientNode(*cn)
	}

	bl := b.Length()

	if bl != len(b.nodes) {
		t.Fatalf("expected length to be %v but got %v", len(b.nodes), bl)
	}
}

package registry_test

import (
	"testing"

	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestNodeMap_Add(t *testing.T) {
	b := registry.NewNodeMap()
	l := 10
	nodes := registry.RemovePointers(registry.CreateTestNodes(l, &registry.TestNodeOptions{
		SubscribedSubjects:  []string{"sub"},
		BroadcastedSubjects: []string{"sub"},
	}))

	for _, n := range nodes {
		b.AddNode(n)
	}

	assert.Len(t, b.Nodes, l)
}

func TestNodeMap_Remove(t *testing.T) {
	b := registry.NewNodeMap()
	l := 10
	nodes := registry.RemovePointers(registry.CreateTestNodes(l, &registry.TestNodeOptions{}))

	for _, n := range nodes {
		b.AddNode(n)
	}

	for _, n := range nodes {
		b.RemoveNode(n.Id)
	}

	assert.Len(t, b.Nodes, 0)
}

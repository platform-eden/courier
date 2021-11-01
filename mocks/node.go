package mocks

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/node"
)

// TestNodeOptions that can be passed into CreateTestNodes
type TestNodeOptions struct {
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}

// CreateTestNodes creates a quantity of randomized nodes based on the options passed in
func CreateTestNodes(count int, options *TestNodeOptions) []*node.Node {
	nodes := []*node.Node{}
	var broadSubjects []string
	var subSubjects []string

	if len(options.SubscribedSubjects) == 0 {
		subSubjects = []string{"sub", "sub1", "sub2"}
	} else {
		subSubjects = options.SubscribedSubjects
	}
	if len(options.BroadcastedSubjects) == 0 {
		broadSubjects = []string{"broad", "broad1"}
	} else {
		broadSubjects = options.BroadcastedSubjects
	}

	for i := 0; i < count; i++ {
		ip := fmt.Sprintf("%v.%v.%v.%v", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
		port := fmt.Sprint(rand.Intn(9999-1000) + 1000)
		subcount := (rand.Intn(len(subSubjects)) + 1)
		broadcount := rand.Intn(len(broadSubjects) + 1)
		var subs []string
		var broads []string

		for i := 0; i < subcount; i++ {
			subs = append(subs, subSubjects[i])
		}

		for i := 0; i < broadcount; i++ {
			broads = append(broads, broadSubjects[i])
		}

		n := node.NewNode(uuid.NewString(), ip, port, subs, broads)
		nodes = append(nodes, n)
	}

	return nodes
}

func RemovePointers(nodes []*node.Node) []node.Node {
	updated := []node.Node{}

	for _, n := range nodes {
		updated = append(updated, *n)
	}

	return updated
}

package courier

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/mocks"
	"github.com/platform-edn/courier/node"
)

/*************************************************************
Expected Outcomes:
- nodes passed into failed connections channel should be added to blacklisted nodes
- nodes passed into failed connections channel should be removed from current nodes
- observe pipeline should be ran every n seconds based on interval
- blacklist and current nodes should be updated after every observe pipeline
*************************************************************/
func TestRegisterNodes(t *testing.T) {
	type test struct {
		blackListCount int
		currentCount   int
		newNodeCount   int
	}

	tests := []test{
		{
			blackListCount: 0,
			currentCount:   0,
			newNodeCount:   5,
		},
		{
			blackListCount: 10,
			currentCount:   10,
			newNodeCount:   25,
		},
		{
			blackListCount: 1000,
			currentCount:   1000,
			newNodeCount:   2500,
		},
	}

	for _, tc := range tests {
		bln := mocks.CreateTestNodes(tc.blackListCount, &mocks.TestNodeOptions{})
		cln := mocks.CreateTestNodes(tc.currentCount, &mocks.TestNodeOptions{})
		nn := mocks.CreateTestNodes(tc.newNodeCount, &mocks.TestNodeOptions{})
		nodes := append(bln, cln...)
		nodes = append(nodes, nn...)
		fn := mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0]
		ochan := make(chan []node.Node)
		nchan := make(chan node.Node)
		fchan := make(chan node.Node)
		blacklist := NewNodeMap()
		current := NewNodeMap()

		go registerNodes(ochan, nchan, fchan, blacklist, current)

		current.Add(*fn)
		fchan <- *fn

		count := 0
		for {
			_, exist1 := blacklist.Node(fn.Id)
			if exist1 {
				_, exist2 := current.Node(fn.Id)
				if !exist2 {
					break
				}
			}
			if count == 3 {
				t.Fatal("expected node to be blacklisted but it wasn't")
			}

			time.Sleep(time.Millisecond * 300)
			count++
		}

		ochan <- mocks.RemovePointers(nodes)
		count = 0

	nodeloop:
		for {
			select {
			case <-nchan:
				count++
				if count == tc.newNodeCount {
					break nodeloop
				}
			case <-time.After(time.Second * 3):
				t.Fatal("did not send all nodes in time")
			}
		}
	}
}

/*************************************************************
Expected Outcomes:
- should return a refreshed list of blacklist and current nodes
- all new nodes should be sent through the new node channel
- no blacklisted nodes should go through the new node channel
- no old nodes should go through the new node channel
*************************************************************/
func TestUpdateNodes(t *testing.T) {
	type test struct {
		blackListCount int
		currentCount   int
		newNodeCount   int
	}

	tests := []test{
		{
			blackListCount: 0,
			currentCount:   0,
			newNodeCount:   5,
		},
		{
			blackListCount: 10,
			currentCount:   10,
			newNodeCount:   25,
		},
		{
			blackListCount: 1000,
			currentCount:   1000,
			newNodeCount:   2500,
		},
	}

	for _, tc := range tests {
		bln := mocks.CreateTestNodes(tc.blackListCount, &mocks.TestNodeOptions{})
		cln := mocks.CreateTestNodes(tc.currentCount, &mocks.TestNodeOptions{})
		nn := mocks.CreateTestNodes(tc.newNodeCount, &mocks.TestNodeOptions{})
		oldNodes := mocks.CreateTestNodes(2, &mocks.TestNodeOptions{})

		blacklist := NewNodeMap()
		current := NewNodeMap()

		for _, n := range bln {
			blacklist.Add(*n)
		}
		blacklist.Add(*oldNodes[0])

		for _, n := range cln {
			current.Add(*n)
		}
		current.Add(*oldNodes[1])

		nodes := append(bln, cln...)
		nodes = append(nodes, nn...)
		nodeChannel := make(chan node.Node, tc.newNodeCount)

		bll, cll := updateNodes(mocks.RemovePointers(nodes), nodeChannel, blacklist, current)
		close(nodeChannel)

		count := 0
		for range nodeChannel {
			count++
		}

		if count != tc.newNodeCount {
			t.Fatalf("expected new node count to be %v but got %v", tc.newNodeCount, count)
		}

		if len(bll) != tc.blackListCount {
			t.Fatalf("expected new blacklist count to be %v but got %v", tc.blackListCount, len(bll))
		}

		if len(cll) != tc.newNodeCount+tc.currentCount {
			t.Fatalf("expected new current count to be %v but got %v", tc.newNodeCount+tc.currentCount, len(cll))
		}

	}
}

/*************************************************************
Expected Outcomes:
- all nodes passed in should be passed out of the out channel
*************************************************************/
func TestGenerate(t *testing.T) {
	type test struct {
		count int
	}

	tests := []test{
		{
			count: 10,
		},
		{
			count: 10000,
		},
		{
			count: 0,
		},
	}

	for _, tc := range tests {
		nodes := mocks.CreateTestNodes(tc.count, &mocks.TestNodeOptions{})

		removePointers := func(nodes []*node.Node) []node.Node {
			updated := []node.Node{}

			for _, n := range nodes {
				updated = append(updated, *n)
			}

			return updated
		}

		out := generate(removePointers(nodes)...)

		count := 0
		for range out {
			count++
		}

		if count != tc.count {
			t.Fatalf("expected %v nodes but got %v nodes", tc.count, count)
		}
	}
}

/*************************************************************
Expected Outcomes:
- if blacklist is empty no nodes should be compared and all should be passed through
- if blacklist is not empty all nodes should be compared to blacklist
- nodes in blacklist should be blacklisted after comparison
- nodes not in blacklist should be passed through
**************************************************************/
func TestCompareBlackList(t *testing.T) {
	type test struct {
		blackListCount int
		newNodeCount   int
	}

	tests := []test{
		{
			blackListCount: 3,
			newNodeCount:   4,
		},
		{
			blackListCount: 0,
			newNodeCount:   3,
		},
		{
			blackListCount: 0,
			newNodeCount:   0,
		},
		{
			blackListCount: 100,
			newNodeCount:   1000,
		},
	}

	for _, tc := range tests {
		pn := mocks.CreateTestNodes(tc.blackListCount, &mocks.TestNodeOptions{})
		blacklist := NewNodeMap()
		for _, n := range pn {
			blacklist.Add(*n)
		}
		oldbln := mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0]
		in := make(chan node.Node)
		blacklisted := make(chan node.Node, tc.blackListCount+tc.newNodeCount)
		newNodes := mocks.CreateTestNodes(tc.newNodeCount, &mocks.TestNodeOptions{})

		blacklist.Add(*oldbln)

		go func() {
			for _, n := range pn {
				in <- *n
			}
			for _, n := range newNodes {
				in <- *n
			}
			close(in)
		}()

		out := compareBlackList(in, blacklisted, blacklist)

		outNodes := []node.Node{}
		count := 0
		for count != tc.newNodeCount {
			select {
			case n := <-out:
				outNodes = append(outNodes, n)
				count++
			case <-time.After(time.Second * 5):
				t.Fatalf("stopped receiving nodes after %v but wanted %v", count, tc.newNodeCount)
			}
		}

		for _, n := range outNodes {
			_, exist := blacklist.Node(n.Id)
			if exist {
				t.Fatalf("node %s was blacklisted but was sent through the out channel", n.Id)
			}
		}

		bc := 0
		for range blacklisted {
			bc++
		}

		if bc != tc.blackListCount {
			t.Fatalf("expected %v blacklist nodes but got %v", tc.blackListCount, bc)
		}

	}
}

/**************************************************************
Expected Outcomes:
- all nodes passed in should be passed into active channel
- all nodes not in current nodes should be passed into out chan
- no nodes in current nodes should be passed into out chan
**************************************************************/
func TestCompareCurrentList(t *testing.T) {
	nc := 5
	nnc := 3
	in := make(chan node.Node)
	active := make(chan node.Node, nc+nnc)
	nodes := mocks.CreateTestNodes(nc, &mocks.TestNodeOptions{})
	newNodes := mocks.CreateTestNodes(nnc, &mocks.TestNodeOptions{})
	current := NewNodeMap()

	for _, n := range nodes {
		current.Add(*n)
	}

	out := compareCurrentList(in, active, current)

	go func() {
		for _, n := range nodes {
			in <- *n
		}
		for _, n := range newNodes {
			in <- *n
		}
		close(in)
	}()

	outNodes := []node.Node{}
	count := 0
	for count != nnc {
		select {
		case n := <-out:
			outNodes = append(outNodes, n)
			count++
		case <-time.After(time.Second * 5):
			t.Fatalf("stopped receiving nodes after %v but wanted %v", count, nnc)
		}
	}

	for _, n := range nodes {
		for _, on := range outNodes {
			if n.Id == on.Id {
				t.Fatalf("out node had a matching id with an already existing node")
			}
		}
	}

	ac := 0
	for range active {
		ac++
	}
	if ac != nc+nnc {
		t.Fatalf("expected total active nodes to be %v but got %v", nc+nnc, ac)
	}
}

/**************************************************************
Expected Outcomes:
- all nodes passed in should be passed into new node channel
- done channel should be sent true after completion
**************************************************************/
func TestSendNodes(t *testing.T) {
	nc := 5
	in := make(chan node.Node)
	new := make(chan node.Node, nc)
	nodes := mocks.CreateTestNodes(nc, &mocks.TestNodeOptions{})

	done := sendNodes(in, new)

	go func() {
		for _, n := range nodes {
			in <- *n
		}

		close(in)
	}()

	d := false

	for !d {
		select {
		case <-done:
			d = true
		case <-time.After(time.Second * 3):
			t.Fatal("did not receive a done signal")
		}
	}
	count := 0
	close(new)

	for range new {
		count++
	}

	if count != nc {
		t.Fatalf("expected receive message count to equal %v but got %v", nc, count)
	}

}

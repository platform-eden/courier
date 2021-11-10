package courier

import (
	"sync"
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
		schan := make(chan node.Node)
		fchan := make(chan node.Node)
		blacklist := NewNodeMap()
		current := NewNodeMap()

		go registerNodes(ochan, nchan, schan, fchan, blacklist, current)

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
		staleNodeCount int
	}

	tests := []test{
		{
			blackListCount: 10000,
			currentCount:   10000,
			newNodeCount:   250000,
			staleNodeCount: 1000,
		},
		{
			blackListCount: 0,
			currentCount:   0,
			newNodeCount:   5,
			staleNodeCount: 3,
		},
		{
			blackListCount: 11,
			currentCount:   10,
			newNodeCount:   25,
			staleNodeCount: 13,
		},
	}

	for _, tc := range tests {
		bln := mocks.CreateTestNodes(tc.blackListCount, &mocks.TestNodeOptions{})
		cln := mocks.CreateTestNodes(tc.currentCount, &mocks.TestNodeOptions{})
		nn := mocks.CreateTestNodes(tc.newNodeCount, &mocks.TestNodeOptions{})
		snl := mocks.CreateTestNodes(tc.staleNodeCount, &mocks.TestNodeOptions{})
		obln := mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0]

		blacklist := NewNodeMap()
		current := NewNodeMap()

		for _, n := range bln {
			blacklist.Add(*n)
		}
		blacklist.Add(*obln)

		for _, n := range cln {
			current.Add(*n)
		}

		for _, n := range snl {
			current.Add(*n)
		}

		nodes := append(bln, cln...)
		nodes = append(nodes, nn...)
		nodeChannel := make(chan node.Node, tc.newNodeCount)
		staleChannel := make(chan node.Node, tc.staleNodeCount)

		bll, cll := updateNodes(mocks.RemovePointers(nodes), nodeChannel, staleChannel, blacklist, current)
		close(nodeChannel)
		close(staleChannel)

		count := 0
		for range nodeChannel {
			count++
		}

		if count != tc.newNodeCount {
			t.Fatalf("expected new node count to be %v but got %v", tc.newNodeCount, count)
		}

		count = 0
		for range staleChannel {
			count++
		}

		if count != tc.staleNodeCount {
			t.Fatalf("expected stale node count to be %v but got %v", tc.staleNodeCount, count)
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
		out := generate(mocks.RemovePointers(nodes)...)

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
- all left nodes in current nods should be removed from current list
- all left over nodes should be sent through stale channel
- no nodes in current nodes should be passed into out chan
**************************************************************/
func TestCompareCurrentList(t *testing.T) {
	type test struct {
		newNodeCount     int
		staleNodeCount   int
		currentNodeCount int
	}

	tests := []test{
		{
			newNodeCount:     10,
			staleNodeCount:   5,
			currentNodeCount: 10,
		},
		{
			newNodeCount:     1000,
			staleNodeCount:   500,
			currentNodeCount: 10000,
		},
	}

	for _, tc := range tests {
		in := make(chan node.Node)
		nodes := mocks.CreateTestNodes(tc.currentNodeCount, &mocks.TestNodeOptions{})
		newNodes := mocks.CreateTestNodes(tc.newNodeCount, &mocks.TestNodeOptions{})
		staleNodes := mocks.CreateTestNodes(tc.staleNodeCount, &mocks.TestNodeOptions{})
		current := NewNodeMap()

		for _, n := range nodes {
			current.Add(*n)
		}

		for _, n := range staleNodes {
			current.Add(*n)
		}

		new, active, stale := compareCurrentList(in, current)

		go func() {
			for _, n := range nodes {
				in <- *n
			}
			for _, n := range newNodes {
				in <- *n
			}
			close(in)
		}()

		type checkTuple struct {
			current  int
			expected int
		}

		check := func(tuples ...checkTuple) bool {
			pass := true
			for _, tuple := range tuples {
				if tuple.current != tuple.expected {
					pass = false
				}
			}

			return pass
		}

		nc := 0
		ac := 0
		sc := 0

		for check(
			checkTuple{
				current:  nc,
				expected: tc.newNodeCount,
			},
			checkTuple{
				current:  ac,
				expected: tc.currentNodeCount + tc.newNodeCount,
			},
			checkTuple{
				current:  sc,
				expected: tc.staleNodeCount,
			},
		) {
			select {
			case <-new:
				nc++
			case <-active:
				ac++
			case <-stale:
				sc++
			case <-time.After(time.Second * 5):
				t.Fatalf("stopped receiving nodes while expecting more: new - %v, active - %v, stale %v", nc, ac, sc)
			}
		}
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
	wg := sync.WaitGroup{}

	wg.Add(1)
	go sendNodes(in, new, &wg)

	go func() {
		for _, n := range nodes {
			in <- *n
		}
		close(in)
	}()

	wg.Wait()

	count := 0
	close(new)
	for range new {
		count++
	}

	if count != nc {
		t.Fatalf("expected receive message count to equal %v but got %v", nc, count)
	}

}

/**************************************************************
Expected Outcomes:
- out channel should be able to hold size amount of nodes
- should unblock processes that don't want to wait for in channel
**************************************************************/
func TestNodeBuffer(t *testing.T) {
	type test struct {
		size int
	}

	tests := []test{
		{
			size: 10,
		},
		{
			size: 1000,
		},
	}

	for _, tc := range tests {
		nodes := mocks.CreateTestNodes(tc.size, &mocks.TestNodeOptions{})
		in := make(chan node.Node)

		out := nodeBuffer(in, tc.size)

		for _, n := range nodes {
			in <- *n
		}
		close(in)

		count := 0
		for range out {
			count++
		}

		if count != tc.size {
			t.Fatalf("expected count to be %v but got %v", tc.size, count)
		}
	}
}

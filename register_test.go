package courier

import (
	"context"
	"sync"
	"testing"
	"time"
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
		bln := CreateTestNodes(tc.blackListCount, &TestNodeOptions{})
		cln := CreateTestNodes(tc.currentCount, &TestNodeOptions{})
		nn := CreateTestNodes(tc.newNodeCount, &TestNodeOptions{})
		nodes := append(bln, cln...)
		nodes = append(nodes, nn...)
		fn := CreateTestNodes(1, &TestNodeOptions{})[0]
		ochan := make(chan []Noder)
		defer close(ochan)
		nchan := make(chan Node, len(nodes))
		defer close(nchan)
		schan := make(chan Node, len(nodes))
		defer close(schan)
		fchan := make(chan Node)
		defer close(fchan)
		blacklist := NewNodeMap(RemovePointers(bln)...)
		current := NewNodeMap(RemovePointers(cln)...)
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		defer wg.Add(1)
		defer cancel()

		go registerNodes(ctx, wg, ochan, nchan, schan, fchan, blacklist, current)

		current.Add(*fn)
		fchan <- *fn

		count := 0
		for {
			_, exist1 := blacklist.Node(fn.id)
			if exist1 {
				_, exist2 := current.Node(fn.id)
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

		noders := []Noder{}
		for _, n := range nodes {
			noders = append(noders, n)
		}

		ochan <- noders
		doneChannel := make(chan struct{})
		go func() {
			for current.Length() != tc.currentCount+tc.newNodeCount {
				time.Sleep(time.Millisecond * 300)
			}

			close(doneChannel)
		}()

		select {
		case <-doneChannel:
			continue
		case <-time.After(time.Second * 3):
			t.Fatal("didn't set current in time")
		}

		wg.Add(1)
		waitChannel := make(chan struct{})
		go func() {
			cancel()
			wg.Wait()
			close(waitChannel)
		}()

		select {
		case <-waitChannel:
			continue
		case <-time.After(time.Second * 3):
			t.Fatal("didn't complete wait group in time")
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
		bln := CreateTestNodes(tc.blackListCount, &TestNodeOptions{})
		cln := CreateTestNodes(tc.currentCount, &TestNodeOptions{})
		nn := CreateTestNodes(tc.newNodeCount, &TestNodeOptions{})
		snl := CreateTestNodes(tc.staleNodeCount, &TestNodeOptions{})
		obln := CreateTestNodes(1, &TestNodeOptions{})[0]

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
		nodeChannel := make(chan Node, tc.newNodeCount)
		staleChannel := make(chan Node, tc.staleNodeCount)
		noders := []Noder{}

		for _, n := range nodes {
			noders = append(noders, n)
		}

		bll, cll := updateNodes(noders, nodeChannel, staleChannel, blacklist, current)
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
- all noders passed in should be passed out of the out channel as nodes
*************************************************************/
func TestGenerateNoders(t *testing.T) {
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
		nodes := CreateTestNodes(tc.count, &TestNodeOptions{})
		noders := []Noder{}

		for _, n := range nodes {
			noders = append(noders, n)
		}

		out := generateNoders(noders...)

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
		pn := CreateTestNodes(tc.blackListCount, &TestNodeOptions{})
		blacklist := NewNodeMap()
		for _, n := range pn {
			blacklist.Add(*n)
		}
		oldbln := CreateTestNodes(1, &TestNodeOptions{})[0]
		in := make(chan Node)
		blacklisted := make(chan Node, tc.blackListCount+tc.newNodeCount)
		newNodes := CreateTestNodes(tc.newNodeCount, &TestNodeOptions{})

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

		outNodes := []Node{}
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
			_, exist := blacklist.Node(n.id)
			if exist {
				t.Fatalf("node %s was blacklisted but was sent through the out channel", n.id)
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
			newNodeCount:     100,
			staleNodeCount:   50,
			currentNodeCount: 1000,
		},
	}

	for _, tc := range tests {
		in := make(chan Node)
		nodes := CreateTestNodes(tc.currentNodeCount, &TestNodeOptions{})
		newNodes := CreateTestNodes(tc.newNodeCount, &TestNodeOptions{})
		staleNodes := CreateTestNodes(tc.staleNodeCount, &TestNodeOptions{})
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
	in := make(chan Node)
	new := make(chan Node, nc)
	nodes := CreateTestNodes(nc, &TestNodeOptions{})
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
			size: 100,
		},
	}

	for _, tc := range tests {
		nodes := CreateTestNodes(tc.size, &TestNodeOptions{})
		in := make(chan Node)

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

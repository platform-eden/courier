package courier

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

/**************************************************************
Expected Outcomes:
- all info passed in will get pushed into responseMap
**************************************************************/
func TestListenForResponseInfo(t *testing.T) {
	type test struct {
		responseCount int
	}

	tests := []test{
		{
			responseCount: 10,
		},
	}

	for _, tc := range tests {
		responses := []ResponseInfo{}
		rchan := make(chan ResponseInfo)
		respMap := newResponseMap()
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()

		for i := 0; i < tc.responseCount; i++ {
			info := ResponseInfo{
				MessageId: uuid.NewString(),
				NodeId:    uuid.NewString(),
			}

			responses = append(responses, info)
		}

		go listenForResponseInfo(ctx, wg, rchan, respMap)

		for _, r := range responses {
			rchan <- r
		}
		defer close(rchan)

		done := make(chan bool)
		go func() {
			for respMap.Length() != tc.responseCount {
				time.Sleep(time.Millisecond * 200)
			}

			done <- true
		}()

		select {
		case <-done:
		case <-time.After(time.Second * 3):
			t.Fatal("nodes not added in time")
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

/**************************************************************
Expected Outcomes:
- all nodes passed in should be turned into clientNodes
- nodes that erred on clientNode creation should be skipped
- all nodes passed in should be added to the subjects they subscribe to in subscribeMap
**************************************************************/
func TestListenForNewNodes(t *testing.T) {
	type test struct {
		newNodes int
		options  []grpc.DialOption
	}

	tests := []test{
		{
			newNodes: 10,
			options:  []grpc.DialOption{grpc.WithInsecure()},
		},
		{
			newNodes: 100,
			options:  []grpc.DialOption{grpc.WithInsecure()},
		},
	}

	for _, tc := range tests {
		nodeMap := newClientNodeMap()
		subMap := newSubscriberMap()
		subjects := []string{"test", "test1", "test2"}
		nodes := CreateTestNodes(tc.newNodes, &TestNodeOptions{SubscribedSubjects: subjects})
		nchan := make(chan Node)
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()

		go listenForNewNodes(ctx, wg, nchan, nodeMap, subMap, uuid.NewString(), tc.options...)

		for _, n := range nodes {
			nchan <- *n
		}
		defer close(nchan)

		done := make(chan bool)

		checkSubMap := func(nodes []*Node, smap *subscriberMap) bool {
			for _, n := range nodes {
				for _, subject := range n.subscribedSubjects {
					exist := smap.CheckForSubscriber(subject, n.id)
					if !exist {
						return false
					}
				}
			}

			return true
		}

		go func() {
			for nodeMap.Length() != tc.newNodes && !checkSubMap(nodes, subMap) {
				time.Sleep(time.Millisecond * 200)
			}

			done <- true
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Millisecond * 300):
			if len(tc.options) == 0 {
				continue
			}
			t.Fatal("nodes not added in time")
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

/**************************************************************
Expected Outcomes:
- all nodes passed in should be removed from
**************************************************************/
func TestListenForStaleNodes(t *testing.T) {
	type test struct {
		staleNodes int
	}

	tests := []test{
		{
			staleNodes: 10,
		},
		{
			staleNodes: 100,
		},
	}

	for _, tc := range tests {
		nodeMap := newClientNodeMap()
		subMap := newSubscriberMap()
		subjects := []string{"test", "test1", "test2"}
		nodes := CreateTestNodes(tc.staleNodes, &TestNodeOptions{SubscribedSubjects: subjects})
		schan := make(chan Node)
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()

		for _, n := range nodes {
			subMap.Add(n.id, n.subscribedSubjects...)
			nodeMap.Add(clientNode{
				Node: *n,
			})
		}

		go listenForStaleNodes(ctx, wg, schan, nodeMap, subMap)

		for _, n := range nodes {
			schan <- *n
		}
		defer close(schan)

		done := make(chan bool)

		checkSubMap := func(nodes []*Node, smap *subscriberMap) bool {
			for _, n := range nodes {
				for _, subject := range n.subscribedSubjects {
					exist := smap.CheckForSubscriber(subject, n.id)
					if exist {
						return false
					}
				}
			}

			return true
		}

		go func() {
			for nodeMap.Length() != 0 && !checkSubMap(nodes, subMap) {
				time.Sleep(time.Millisecond * 200)
			}

			done <- true
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Second * 3):
			t.Fatal("nodes not removed from nodeMap in time")
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

/**************************************************************
Expected Outcomes:
- should take in a subject and find all node ids subscribed to that subject
- returns an error if subject doesn't exist
- all node ids subscribed to that subject are sent through the out channel
**************************************************************/
func TestGenerateIdsBySubject(t *testing.T) {
	type test struct {
		expectedFailure bool
		nodeCount       int
	}

	tests := []test{
		{
			expectedFailure: false,
			nodeCount:       10,
		},
		{
			expectedFailure: true,
			nodeCount:       10,
		},
	}

	for _, tc := range tests {
		subMap := newSubscriberMap()
		subject := "test"
		for i := 0; i < tc.nodeCount; i++ {
			id := uuid.NewString()
			subMap.Add(id, subject)
		}

		if tc.expectedFailure {
			_, err := generateIdsBySubject("fail", subMap)
			if err != nil {
				break
			} else {
				t.Fatal("expected generateIdsBySubject to fail but it didn't")
			}
		}

		out, err := generateIdsBySubject(subject, subMap)
		if err != nil {
			t.Fatalf("expected generateIdsBySubject to pass but it failed: %s", err)
		}

		count := 0
		for range out {
			count++
		}

		if count != tc.nodeCount {
			t.Fatalf("expected out count to be %v but got %v", tc.nodeCount, count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- should return a channel of string holding nodeIds
- should return an error if responseMap doesn't have a messageId that matches the given one
**************************************************************/
func TestGenerateIdsByMessage(t *testing.T) {
	type test struct {
		messageCount    int
		expectedFailure bool
	}

	tests := []test{
		{
			messageCount:    1,
			expectedFailure: false,
		},
		{
			messageCount:    1,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		respMap := newResponseMap()
		messages := []string{}

		for i := 0; i < tc.messageCount; i++ {
			info := ResponseInfo{
				MessageId: uuid.NewString(),
				NodeId:    uuid.NewString(),
			}

			messages = append(messages, info.MessageId)
			respMap.Push(info)
		}

		var out <-chan string
		var err error
		if tc.expectedFailure {
			out, err = generateIdsByMessage(uuid.NewString(), respMap)
		} else {
			out, err = generateIdsByMessage(messages[0], respMap)
		}

		if err != nil && !tc.expectedFailure {
			t.Fatalf("expected to pass but it failed: %s", err)
		}

		if tc.expectedFailure {
			break
		}

		count := 0
		for range out {
			count++
		}

		if count != 1 {
			t.Fatalf("expected count to equal 1 but got %v", count)
		}

	}
}

/**************************************************************
Expected Outcomes:
- all ids passed in with a valid node will have a clientNode passed out
- all ids that don't have a valid node wil be skipped
**************************************************************/
func TestIdToClientNodes(t *testing.T) {
	type test struct {
		nodeCount    int
		invalidCount int
	}

	tests := []test{
		{
			nodeCount:    10,
			invalidCount: 5,
		},
		{
			nodeCount:    100,
			invalidCount: 50,
		},
	}

	for _, tc := range tests {
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		nodeMap := newClientNodeMap()
		in := make(chan string)
		ids := []string{}
		for _, n := range nodes {
			cn, err := newClientNode(*n, uuid.NewString(), grpc.WithInsecure())
			if err != nil {
				t.Fatalf("could not create clientNode: %s", err)
			}
			nodeMap.Add(*cn)
			ids = append(ids, n.id)
		}

		for i := 0; i < tc.invalidCount; i++ {
			ids = append(ids, uuid.NewString())
		}

		out := idToClientNodes(in, nodeMap)

		go func() {
			for _, id := range ids {
				in <- id
			}
			close(in)
		}()

		count := 0
		for range out {
			count++
		}

		if count != tc.nodeCount {
			t.Fatalf("expected %v nodes to be output but got %v instead", tc.nodeCount, count)
		}

	}
}

/**************************************************************
Expected Outcomes:
- all courierClients passed in should have a message attempted
- all failing send attempts should pass the node they failed on through the out channel
- out channel should close once all courierClients are passed through
**************************************************************/
func TestFanMessageAttempts(t *testing.T) {
	type test struct {
		failures    int
		clientCount int
	}

	tests := []test{
		{
			failures:    3,
			clientCount: 9,
		},
		{
			failures:    75,
			clientCount: 100,
		},
		{
			failures:    500,
			clientCount: 1100,
		},
	}

	for _, tc := range tests {
		metadata := attemptMetadata{
			maxAttempts:  1,
			waitInterval: time.Millisecond * 100,
		}
		ctx := context.Background()
		msg := NewPubMessage("test", "test", []byte("test"))
		nodes := CreateTestNodes(tc.clientCount, &TestNodeOptions{})
		in := make(chan Sender, tc.clientCount)

		fails := 0
		for _, n := range nodes {
			clientFail := false
			if fails < tc.failures {
				clientFail = true
			}
			fails++

			c := NewMockClientNode(*n, clientFail)

			in <- c
		}

		out := fanMessageAttempts(in, ctx, metadata, msg)

		close(in)
		count := 0
		for range out {
			count++
		}

		if count != tc.failures {
			t.Fatalf("expected %v amount of failures but got %v instead", tc.failures, count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- should complete withouot passing anything through the nchan on successful send
- should send the node that it failed to connect to through the nchan
- should mark wait group as complete when finished whether passing or failing
**************************************************************/
func TestAttemptMessage(t *testing.T) {
	type test struct {
		expectedFailure bool
	}

	tests := []test{
		{
			expectedFailure: true,
		},
		{
			expectedFailure: false,
		},
	}

	for _, tc := range tests {
		wg := &sync.WaitGroup{}
		metadata := attemptMetadata{
			maxAttempts:  3,
			waitInterval: time.Millisecond * 100,
		}
		ctx := context.Background()
		msg := NewPubMessage("test", "test", []byte("test"))
		nchan := make(chan Node, 1)
		client := NewMockClientNode(*CreateTestNodes(1, &TestNodeOptions{})[0], tc.expectedFailure)
		wg.Add(1)

		attemptMessage(ctx, client, metadata, msg, nchan, wg)

		wg.Wait()

		close(nchan)
		count := 0
		for range nchan {
			count++
		}

		if tc.expectedFailure && count == 0 {
			t.Fatal("expected failed node count to be 1 but got 0")
		}

		if !tc.expectedFailure && count > 0 {
			t.Fatalf("expected failed node count to be 0 but got %v", count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- all messages passed in should be passed out the fchan and schan
- messages passed through the fchan should be the same as the schan
**************************************************************/
func TestForwardFailedMessages(t *testing.T) {
	type test struct {
		nodeCount int
	}

	tests := []test{
		{
			nodeCount: 10,
		},
		{
			nodeCount: 0,
		},
		{
			nodeCount: 100,
		},
	}

	for _, tc := range tests {
		schan := make(chan Node, tc.nodeCount)
		fchan := make(chan Node, tc.nodeCount)
		in := make(chan Node)
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})

		out := forwardFailedConnections(in, fchan, schan)

		go func() {
			for _, n := range nodes {
				in <- *n
			}

			close(in)
		}()

		select {
		case <-out:
		case <-time.After(time.Second * 3):
			t.Fatal("did not receive done")
		}

		close(schan)
		close(fchan)

		fcount := 0
		for range fchan {
			fcount++
		}

		scount := 0
		for range schan {
			scount++
		}

		if scount != tc.nodeCount || fcount != tc.nodeCount {
			t.Fatalf("mismatched counts: expected: %v, fchan: %v, schan: %v", tc.nodeCount, fcount, scount)
		}

	}
}

package messaging

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func TestNodeIdGenerationError_Error(t *testing.T) {
	method := "testMethod"
	err := fmt.Errorf("test error")
	e := &NodeIdGenerationError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

func TestMessagingClient_Stop(t *testing.T) {
	fchan := make(chan Node)
	schan := make(chan Node)
	nchan := make(chan Node)
	rchan := make(chan ResponseInfo)

	client := newMessagingClient(&messageClientOptions{
		failedChannel:   fchan,
		staleChannel:    schan,
		nodeChannel:     nchan,
		responseChannel: rchan,
		currentId:       "testId",
		clientOptions:   []ClientNodeOption{},
		startClient:     true,
	})

	client.stop()
}

func TestMessagingClient_Publish(t *testing.T) {
	type test struct {
		nodeCount       int
		nodeSubject     string
		messageSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			nodeSubject:     "test",
			messageSubject:  "test",
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		if tc.expectedFailure {
			testMessageServer.SetToFail()
		} else {
			testMessageServer.SetToPass()
		}
		testMessageServer.Clear()

		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		fchan := make(chan Node)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     true,
		})

		for _, n := range nodes {
			client.subscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
			if err != nil {
				t.Fatalf("could not create grpc client: %s", err)
			}

			cn := clientNode{
				Node:       *n,
				connection: conn,
				currentId:  client.currentId,
			}

			client.clientNodes.Add(cn)
		}

		done := make(chan struct{})
		errchan := make(chan error)
		go func() {
			msg := NewPubMessage(uuid.NewString(), tc.messageSubject, []byte("test"))
			err := client.publish(context.Background(), msg)
			if err != nil {
				errchan <- err
			}

			close(done)
		}()

	doneloop:
		for {
			select {
			case <-done:
				if !tc.expectedFailure {
					if testMessageServer.MessagesLength() != tc.nodeCount {
						t.Fatalf("expected %v messages to be sent but got %v instead", tc.nodeCount, testMessageServer.MessagesLength())
					}
				}
				break doneloop
			case err := <-errchan:
				if !tc.expectedFailure {
					client.stop()
					close(errchan)
					t.Fatalf("could not request message: %s", err)
				}
				continue
			case <-time.After(time.Second * 3):
				t.Fatalf("did not send Publish in time")
			}
		}
		client.stop()
		close(errchan)
	}
}

func TestMessagingClient_Request(t *testing.T) {
	defer testMessageServer.SetToPass()

	type test struct {
		nodeCount       int
		nodeSubject     string
		messageSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			nodeSubject:     "test",
			messageSubject:  "test",
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		if tc.expectedFailure {
			testMessageServer.SetToFail()
		} else {
			testMessageServer.SetToPass()
		}

		testMessageServer.Clear()

		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		fchan := make(chan Node)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     true,
		})

		for _, n := range nodes {
			client.subscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
			if err != nil {
				t.Fatalf("could not create grpc client: %s", err)
			}

			cn := clientNode{
				Node:       *n,
				connection: conn,
				currentId:  client.currentId,
			}

			client.clientNodes.Add(cn)
		}

		done := make(chan struct{})
		errchan := make(chan error)
		go func() {
			msg := NewReqMessage(uuid.NewString(), tc.messageSubject, []byte("test"))
			err := client.request(context.Background(), msg)
			if err != nil {
				errchan <- err
			}

			close(done)
		}()

	doneloop:
		for {
			select {
			case <-done:
				if !tc.expectedFailure {
					if testMessageServer.MessagesLength() != tc.nodeCount {
						t.Fatalf("expected %v messages to be sent but got %v instead", tc.nodeCount, testMessageServer.MessagesLength())
					}

					if testMessageServer.ResponsesLength() != tc.nodeCount {
						t.Fatalf("expected %v responses to be sent but got %v instead", tc.nodeCount, testMessageServer.ResponsesLength())
					}
				}
				break doneloop
			case err := <-errchan:
				if !tc.expectedFailure {
					client.stop()
					close(errchan)
					t.Fatalf("could not request message: %s", err)
				}
				continue
			case <-time.After(time.Second * 3):
				t.Fatalf("did not send Response in time")
			}
		}
		client.stop()
		close(errchan)
	}
}

func TestMessagingClient_Response(t *testing.T) {
	defer testMessageServer.SetToPass()

	type test struct {
		nodeCount       int
		subject         string
		badMessageId    bool
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			subject:         "test",
			badMessageId:    false,
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			subject:         "test",
			badMessageId:    true,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		if tc.expectedFailure {
			testMessageServer.SetToFail()
		} else {
			testMessageServer.SetToPass()
		}
		testMessageServer.Clear()

		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		fchan := make(chan Node)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     true,
		})

		messageIds := []string{}
		_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
		if err != nil {
			t.Fatalf("could not create mock client: %s", err)
		}

		for _, node := range nodes {
			cn := clientNode{
				Node:       *node,
				connection: conn,
				currentId:  client.currentId,
			}

			client.clientNodes.Add(cn)

			id := uuid.NewString()
			messageIds = append(messageIds, id)

			if tc.badMessageId {
				client.responses.Push(ResponseInfo{
					NodeId:    node.id,
					MessageId: "badId",
				})
			} else {
				client.responses.Push(ResponseInfo{
					NodeId:    node.id,
					MessageId: id,
				})
			}
		}

		errchan := make(chan error)
		done := make(chan struct{})

		go func() {
			ctx := context.Background()
			for _, id := range messageIds {
				msg := NewRespMessage(id, tc.subject, []byte("test"))
				err := client.response(ctx, msg)
				if err != nil {
					errchan <- err
				}
			}

			close(done)
		}()

	doneloop:
		for {
			select {
			case <-done:
				if !tc.expectedFailure {
					if testMessageServer.MessagesLength() != tc.nodeCount {
						t.Fatalf("expected messages sent to be %v but got %v", tc.nodeCount, testMessageServer.MessagesLength())
					}
				}
				break doneloop
			case err := <-errchan:
				if !tc.expectedFailure {
					client.stop()
					close(errchan)
					t.Fatalf("expected test to pass but it failed sending response: %s", err)
				}
			case <-time.After(time.Second * 10):
				t.Fatal("did not finish sending responses in time")
			}
		}

		client.stop()
		close(errchan)
	}
}

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
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()
		fchan := make(chan Node)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     false,
		})

		go client.listenForResponseInfo(ctx, wg)

		for i := 0; i < tc.responseCount; i++ {
			info := ResponseInfo{
				MessageId: uuid.NewString(),
				NodeId:    uuid.NewString(),
			}

			responses = append(responses, info)
		}

		for _, r := range responses {
			rchan <- r
		}

		defer close(rchan)

		done := make(chan bool)
		go func() {
			for client.responses.Length() != tc.responseCount {
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

		close(fchan)
		close(schan)
		close(rchan)
		close(nchan)
	}
}

func TestListenForNewNodes(t *testing.T) {
	type test struct {
		newNodes int
	}

	tests := []test{
		{
			newNodes: 10,
		},
		{
			newNodes: 100,
		},
	}

	for _, tc := range tests {
		subjects := []string{"test", "test1", "test2"}
		nodes := CreateTestNodes(tc.newNodes, &TestNodeOptions{SubscribedSubjects: subjects})
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()
		fchan := make(chan Node)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     false,
		})

		go client.listenForNewNodes(ctx, wg)

		for _, n := range nodes {
			nchan <- *n
		}
		defer close(nchan)

		done := make(chan bool)

		checkSubMap := func(nodes []*Node, smap SubMapper) bool {
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
			for client.clientNodes.Length() != tc.newNodes && !checkSubMap(nodes, client.subscribers) {
				time.Sleep(time.Millisecond * 200)
			}

			done <- true
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Millisecond * 300):
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

		close(fchan)
		close(schan)
		close(rchan)
		close(nchan)
	}
}

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
		subjects := []string{"test", "test1", "test2"}
		nodes := CreateTestNodes(tc.staleNodes, &TestNodeOptions{SubscribedSubjects: subjects})
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer wg.Add(1)
		defer cancel()
		fchan := make(chan Node, tc.staleNodes)
		schan := make(chan Node)
		nchan := make(chan Node)
		rchan := make(chan ResponseInfo)

		client := newMessagingClient(&messageClientOptions{
			failedChannel:   fchan,
			staleChannel:    schan,
			nodeChannel:     nchan,
			responseChannel: rchan,
			currentId:       "testId",
			clientOptions:   []ClientNodeOption{},
			startClient:     false,
		})

		for _, n := range nodes {
			client.subscribers.Add(n.id, n.subscribedSubjects...)
			client.clientNodes.Add(clientNode{
				Node: *n,
			})
		}

		go client.listenForStaleNodes(ctx, wg)

		for _, n := range nodes {
			schan <- *n
		}
		defer close(schan)

		done := make(chan bool)

		checkSubMap := func(nodes []*Node, smap SubMapper) bool {
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
			for client.clientNodes.Length() != 0 && !checkSubMap(nodes, client.subscribers) {
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

		close(fchan)
		close(schan)
		close(rchan)
		close(nchan)
	}
}

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
			cn, err := newClientNode(*n, uuid.NewString(), WithDialOptions(grpc.WithInsecure()))
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

		out := fanMessageAttempts(in, ctx, msg)

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
		ctx := context.Background()
		msg := NewPubMessage("test", "test", []byte("test"))
		nchan := make(chan Node, 1)
		client := NewMockClientNode(*CreateTestNodes(1, &TestNodeOptions{})[0], tc.expectedFailure)
		wg.Add(1)

		attemptMessage(ctx, client, msg, nchan, wg)

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
		subscribers := newSubscriberMap()
		clientNodes := newClientNodeMap()
		fchan := make(chan Node, tc.nodeCount)
		in := make(chan Node)
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})

		out := forwardFailedConnections(in, clientNodes, subscribers, fchan)

		for _, n := range nodes {
			n.subscribedSubjects = []string{"test"}
			subscribers.Add(n.id, "test")
			clientNodes.Add(clientNode{
				Node: *n,
			})
		}

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

		close(fchan)
		fcount := 0
		for range fchan {
			fcount++
		}

		if fcount != tc.nodeCount {
			t.Fatalf("expected %v failed nodes but got %v", tc.nodeCount, fcount)
		}

		for _, n := range nodes {
			if subscribers.CheckForSubscriber("test", n.id) {
				t.Fatalf("subscriber %s was not removed", n.id)
			}
		}

		if clientNodes.Length() != 0 {
			t.Fatalf("expected clientNode length to be 0 but got %v", clientNodes.Length())
		}
	}
}

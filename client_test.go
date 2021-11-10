package courier

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/mocks"
	"github.com/platform-edn/courier/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
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
		responses := []node.ResponseInfo{}
		rchan := make(chan node.ResponseInfo)
		respMap := newResponseMap()

		for i := 0; i < tc.responseCount; i++ {
			info := node.ResponseInfo{
				MessageId: uuid.NewString(),
				NodeId:    uuid.NewString(),
			}

			responses = append(responses, info)
		}

		go listenForResponseInfo(rchan, respMap)

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

	}
}

/**************************************************************
Expected Outcomes:
- all nodes passed in should be added to the nodeMap
- all nodes passed in should be added to the subjects they subscribe to in subscribeMap
**************************************************************/
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
		nodeMap := NewNodeMap()
		subMap := newSubscriberMap()
		subjects := []string{"test", "test1", "test2"}
		nodes := mocks.CreateTestNodes(tc.newNodes, &mocks.TestNodeOptions{SubscribedSubjects: subjects})
		nchan := make(chan node.Node)

		go listenForNewNodes(nchan, nodeMap, subMap)

		for _, n := range nodes {
			nchan <- *n
		}
		defer close(nchan)

		done := make(chan bool)

		checkSubMap := func(nodes []*node.Node, smap *subscriberMap) bool {
			for _, n := range nodes {
				for _, subject := range n.SubscribedSubjects {
					exist := smap.CheckForSubscriber(subject, n.Id)
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
		case <-time.After(time.Second * 3):
			t.Fatal("nodes not added in time")
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
		nodeMap := NewNodeMap()
		subMap := newSubscriberMap()
		subjects := []string{"test", "test1", "test2"}
		nodes := mocks.CreateTestNodes(tc.staleNodes, &mocks.TestNodeOptions{SubscribedSubjects: subjects})
		schan := make(chan node.Node)

		for _, n := range nodes {
			subMap.Add(n.Id, n.SubscribedSubjects...)
			nodeMap.Add(*n)
		}

		go listenForStaleNodes(schan, nodeMap, subMap)

		for _, n := range nodes {
			schan <- *n
		}
		defer close(schan)

		done := make(chan bool)

		checkSubMap := func(nodes []*node.Node, smap *subscriberMap) bool {
			for _, n := range nodes {
				for _, subject := range n.SubscribedSubjects {
					exist := smap.CheckForSubscriber(subject, n.Id)
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
			info := node.ResponseInfo{
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
- all ids passed in with a valid node will have a node passed out
- all ids that don't have a valid node wil be skipped
**************************************************************/
func TestIdToNodes(t *testing.T) {
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
		nodes := mocks.CreateTestNodes(tc.nodeCount, &mocks.TestNodeOptions{})
		nodeMap := NewNodeMap()
		in := make(chan string)
		ids := []string{}
		for _, n := range nodes {
			nodeMap.Add(*n)
			ids = append(ids, n.Id)
		}

		for i := 0; i < tc.invalidCount; i++ {
			ids = append(ids, uuid.NewString())
		}

		out := idToNodes(in, nodeMap)

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
- every node in should have a courierClient sent through the out channel
**************************************************************/
func TestNodeToCourierClients(t *testing.T) {
	type test struct {
		nodeCount int
	}

	tests := []test{
		{
			nodeCount: 10,
		},
		{
			nodeCount: 1000,
		},
	}

	for _, tc := range tests {
		nodes := mocks.CreateTestNodes(int(tc.nodeCount), &mocks.TestNodeOptions{})
		in := make(chan node.Node)

		go func() {
			for _, n := range nodes {
				in <- *n
			}
			close(in)
		}()

		out := nodeToCourierClients(in, uuid.NewString(), grpc.WithInsecure())

		count := 0
		for range out {
			count++
		}

		if count != tc.nodeCount {
			t.Fatalf("expected %v clients but got %v", tc.nodeCount, count)
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
		msg := message.NewPubMessage("test", "test", []byte("test"))
		nodes := mocks.CreateTestNodes(tc.clientCount, &mocks.TestNodeOptions{})
		in := make(chan courierClient, tc.clientCount)
		// clients := {}
		for _, n := range nodes {
			c := courierClient{
				receiver: *n,
			}

			in <- c
		}

		fcount := 0
		scount := 0
		l := lock.NewTicketLock()
		send := func(ctx context.Context, msg message.Message, client courierClient) error {
			fail := false
			l.Lock()
			scount++
			if fcount != tc.failures {
				fcount++
				fail = true
			}
			l.Unlock()

			if fail {
				return fmt.Errorf("failure!")
			}

			return nil
		}

		out := fanMessageAttempts(in, ctx, metadata, msg, send)

		close(in)
		count := 0
		for range out {
			count++
		}

		if count != tc.failures {
			t.Fatalf("expected %v amount of failures but got %v instead", tc.failures, count)
		}

		if scount != tc.clientCount {
			t.Fatalf("expected %v amount of sends but got %v instead", tc.clientCount, scount)
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
		msg := message.NewPubMessage("test", "test", []byte("test"))
		nchan := make(chan node.Node, 1)
		client := courierClient{
			receiver: *mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0],
		}
		send := func(ctx context.Context, msg message.Message, client courierClient) error {
			if tc.expectedFailure == true {
				return fmt.Errorf("failure!")
			}

			return nil
		}

		wg.Add(1)

		attemptMessage(ctx, client, metadata, msg, send, nchan, wg)

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
			nodeCount: 1000,
		},
	}

	for _, tc := range tests {
		schan := make(chan node.Node, tc.nodeCount)
		fchan := make(chan node.Node, tc.nodeCount)
		in := make(chan node.Node)
		nodes := mocks.CreateTestNodes(tc.nodeCount, &mocks.TestNodeOptions{})

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

/**************************************************************
Expected Outcomes:
- should send a publish message to a grpc server successfully
- returns error if the message was not sent successfully
- returns error if message is not a publish message
**************************************************************/
func TestSendPublishMessage(t *testing.T) {
	type test struct {
		m               message.Message
		serverFailure   bool
		expectedFailure bool
	}

	tests := []test{
		{
			m:               message.NewPubMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   true,
			expectedFailure: true,
		},
		{
			m:               message.NewPubMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: false,
		},
		{
			m:               message.NewReqMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := mocks.NewMockServer(bufconn.Listen(1024*1024), tc.serverFailure)
		client, conn, err := mocks.NewLocalGRPCClient("bufnet", server.BufDialer)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not creat client for server: %s", err)
		}
		defer conn.Close()

		cc := courierClient{
			client:     client,
			connection: *conn,
			currentId:  uuid.NewString(),
			receiver:   *mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0],
		}

		err = sendPublishMessage(context.Background(), tc.m, cc)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not send message: %s", err)
		}

		if tc.expectedFailure {
			t.Fatalf("sendPublishMessage was expected to fail but it didn't")
		}
	}
}

/**************************************************************
Expected Outcomes:
- should send a request message to a grpc server successfully
- returns error if the message was not sent successfully
- returns error if message is not a publish message
**************************************************************/
func TestSendRequestMessage(t *testing.T) {
	type test struct {
		m               message.Message
		serverFailure   bool
		expectedFailure bool
	}

	tests := []test{
		{
			m:               message.NewReqMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   true,
			expectedFailure: true,
		},
		{
			m:               message.NewReqMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: false,
		},
		{
			m:               message.NewPubMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := mocks.NewMockServer(bufconn.Listen(1024*1024), tc.serverFailure)
		client, conn, err := mocks.NewLocalGRPCClient("bufnet", server.BufDialer)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not creat client for server: %s", err)
		}
		defer conn.Close()

		cc := courierClient{
			client:     client,
			connection: *conn,
			currentId:  uuid.NewString(),
			receiver:   *mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0],
		}

		err = sendRequestMessage(context.Background(), tc.m, cc)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not send message: %s", err)
		}

		if tc.expectedFailure {
			t.Fatalf("sendRequestMessage was expected to fail but it didn't")
		}
	}
}

/**************************************************************
Expected Outcomes:
- should send a request message to a grpc server successfully
- returns error if the message was not sent successfully
- returns error if message is not a publish message
**************************************************************/
func TestSendResponseMessage(t *testing.T) {
	type test struct {
		m               message.Message
		serverFailure   bool
		expectedFailure bool
	}

	tests := []test{
		{
			m:               message.NewRespMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   true,
			expectedFailure: true,
		},
		{
			m:               message.NewRespMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: false,
		},
		{
			m:               message.NewPubMessage(uuid.NewString(), "test", []byte("test")),
			serverFailure:   false,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := mocks.NewMockServer(bufconn.Listen(1024*1024), tc.serverFailure)
		client, conn, err := mocks.NewLocalGRPCClient("bufnet", server.BufDialer)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not creat client for server: %s", err)
		}
		defer conn.Close()

		cc := courierClient{
			client:     client,
			connection: *conn,
			currentId:  uuid.NewString(),
			receiver:   *mocks.CreateTestNodes(1, &mocks.TestNodeOptions{})[0],
		}

		err = sendResponseMessage(context.Background(), tc.m, cc)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("could not send message: %s", err)
		}

		if tc.expectedFailure {
			t.Fatalf("sendResponseMessage was expected to fail but it didn't")
		}
	}
}

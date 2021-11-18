package courier

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

/**************************************************************
Expected Outcomes:
- returns a courier object with all passed in options
- returns an error if an observer option isn't passed
- assigns a local ip if a hostname isn't passed
- assigns WithInsecure if a DialOption isn't passed
**************************************************************/
func TestNewCourier(t *testing.T) {
	type test struct {
		observer        Observer
		dialOptions     []grpc.DialOption
		hostname        string
		port            string
		startOnCreation bool
		expectedFailure bool
	}

	tests := []test{
		{
			observer:        newMockObserver(make(chan []Noder), false),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
			port:            "3000",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        nil,
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
			port:            "3001",
			startOnCreation: false,
			expectedFailure: true,
		},
		{
			observer:        newMockObserver(make(chan []Noder), false),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "",
			port:            "3002",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        newMockObserver(make(chan []Noder), false),
			dialOptions:     []grpc.DialOption{},
			hostname:        "test",
			port:            "3003",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        newMockObserver(make(chan []Noder), true),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
			port:            "3004",
			startOnCreation: true,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		sub := []string{"sub1", "sub2", "sub3"}
		broad := []string{"broad1", "broad2", "broad3"}

		c, err := NewCourier(
			WithObserver(tc.observer),
			Subscribes(sub...),
			Broadcasts(broad...),
			WithHostname(tc.hostname),
			WithPort(tc.port),
			WithDialOptions(tc.dialOptions...),
			WithFailedMessageWaitInterval(time.Second),
			WithMaxFailedMessageAttempts(5),
			StartOnCreation(tc.startOnCreation),
		)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("expected NewCourier to pass but it failed: %s", err)
		}

		if c.Hostname == "" {
			t.Fatal("Hostname should not be blank but it is")
		}
		if c.Observer == nil {
			t.Fatal("Observer should not be blank but it is")
		}
		if len(c.DialOptions) == 0 {
			t.Fatal("DialOptions length should not be 0 but it is")
		}

		if tc.expectedFailure {
			t.Fatal("expected test to fail but it passed")
		}

		c.Stop()
	}
}

/**************************************************************
Expected Outcomes:
- errors when port is not a valid port
- errors if Observer can't update node
- returns nil if server and go routines start
**************************************************************/
func TestCourier_Start(t *testing.T) {
	type test struct {
		port            string
		observer        Observer
		expectedFailure bool
	}

	tests := []test{
		{
			port:            "asdfasdf",
			observer:        newMockObserver(make(chan []Noder), false),
			expectedFailure: true,
		},
		{
			port:            "3000",
			observer:        newMockObserver(make(chan []Noder), true),
			expectedFailure: true,
		},
		{
			port:            "3001",
			observer:        newMockObserver(make(chan []Noder), false),
			expectedFailure: false,
		},
	}

	for _, tc := range tests {
		sub := []string{"sub1", "sub2", "sub3"}
		broad := []string{"broad1", "broad2", "broad3"}

		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			Subscribes(sub...),
			Broadcasts(broad...),
			WithHostname("test.com"),
			WithPort(tc.port),
			WithDialOptions([]grpc.DialOption{grpc.WithInsecure()}...),
			WithFailedMessageWaitInterval(time.Second),
			WithMaxFailedMessageAttempts(5),
			StartOnCreation(false),
		)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("expected NewCourier to pass but it failed: %s", err)
		}

		err = c.Start()
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("expected Start to pass but it failed: %s", err)
		}

		c.Stop()
	}
}

type createMessage func(string, string, []byte) Message

/**************************************************************
Expected Outcomes:
- should return an error if the message passed doesn't have a topic registered by client nodes
- should send messages to all clients subscribed to the subject
**************************************************************/
func TestCourier_Publish(t *testing.T) {
	type test struct {
		nodeCount       int
		port            string
		nodeSubject     string
		messageSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "test",
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := NewMockServer(bufconn.Listen(1024*1024), false)
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			WithPort(tc.port),
			WithDialOptions(grpc.WithInsecure()),
			WithFailedMessageWaitInterval(time.Second),
			WithMaxFailedMessageAttempts(5),
			StartOnCreation(false),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		for _, n := range nodes {
			c.clientSubscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewLocalGRPCClient("bufnet", server.BufDialer)
			if err != nil {
				t.Fatalf("could not create grpc client: %s", err)
			}

			cn := clientNode{
				Node:       *n,
				connection: conn,
				currentId:  c.Id,
			}

			c.clientNodes.Add(cn)
		}

		done := make(chan bool)
		errchan := make(chan error)
		go func() {
			err := c.Publish(context.Background(), tc.nodeSubject, []byte("test"))
			if err != nil {
				errchan <- err
				return
			}

			done <- true
		}()

		select {
		case <-done:
		case err := <-errchan:
			if !tc.expectedFailure {
				t.Fatalf("could not publish message: %s", err)
			}
			continue
		case <-time.After(time.Second * 3):
			t.Fatalf("did not Publish in time")
		}

		if server.MessagesLength() != tc.nodeCount {
			t.Fatalf("expected %v messages to be sent but got %v instead", tc.nodeCount, server.MessagesLength())
		}
	}
}

/**************************************************************
Expected Outcomes:
- should return an error if the message passed doesn't have a topic registered by client nodes
- should send messages to all clients subscribed to the subject
**************************************************************/
func TestCourier_Request(t *testing.T) {
	type test struct {
		nodeCount       int
		port            string
		nodeSubject     string
		messageSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "test",
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			messageSubject:  "fail",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := NewMockServer(bufconn.Listen(1024*1024), false)
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			WithPort(tc.port),
			WithDialOptions(grpc.WithInsecure()),
			WithFailedMessageWaitInterval(time.Second),
			WithMaxFailedMessageAttempts(5),
			StartOnCreation(false),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		for _, n := range nodes {
			c.clientSubscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewLocalGRPCClient("bufnet", server.BufDialer)
			if err != nil {
				t.Fatalf("could not create grpc client: %s", err)
			}

			cn := clientNode{
				Node:       *n,
				connection: conn,
				currentId:  c.Id,
			}

			c.clientNodes.Add(cn)
		}

		done := make(chan bool)
		errchan := make(chan error)
		go func() {
			err := c.Request(context.Background(), tc.messageSubject, []byte("test"))
			if err != nil {
				errchan <- err
				return
			}

			done <- true
		}()

		select {
		case <-done:
		case err := <-errchan:
			if !tc.expectedFailure {
				t.Fatalf("could not request message: %s", err)
			}
			continue
		case <-time.After(time.Second * 3):
			t.Fatalf("did not Request in time")
		}

		if server.MessagesLength() != tc.nodeCount {
			t.Fatalf("expected %v messages to be sent but got %v instead", tc.nodeCount, server.MessagesLength())
		}

		if server.ResponsesLength() != tc.nodeCount {
			t.Fatalf("expected %v responses to be sent but got %v instead", tc.nodeCount, server.ResponsesLength())
		}
	}
}

/**************************************************************
Expected Outcomes:
- returns error if messageId isn't in responseMap
- sends a message to the node in the responseInfo
**************************************************************/
func TestCourier_Response(t *testing.T) {
	type test struct {
		nodeCount       int
		port            string
		nodeSubject     string
		badMessageId    bool
		expectedFailure bool
	}

	tests := []test{
		{
			nodeCount:       5,
			port:            "3001",
			nodeSubject:     "test",
			badMessageId:    false,
			expectedFailure: false,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			badMessageId:    true,
			expectedFailure: true,
		},
		{
			nodeCount:       1,
			port:            "3001",
			nodeSubject:     "test",
			badMessageId:    false,
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		server := NewMockServer(bufconn.Listen(1024*1024), false)
		nodes := CreateTestNodes(tc.nodeCount, &TestNodeOptions{})
		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			WithPort(tc.port),
			WithDialOptions(grpc.WithInsecure()),
			WithFailedMessageWaitInterval(time.Second),
			WithMaxFailedMessageAttempts(5),
			StartOnCreation(false),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		var messageList []Message

		for _, n := range nodes {
			c.clientSubscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewLocalGRPCClient("bufnet", server.BufDialer)
			if err != nil {
				t.Fatalf("could not create grpc client: %s", err)
			}

			cn := clientNode{
				Node:       *n,
				connection: conn,
				currentId:  c.Id,
			}

			c.clientNodes.Add(cn)

			m := NewRespMessage(uuid.NewString(), tc.nodeSubject, []byte("test"))

			if tc.badMessageId {
				c.responses.Push(ResponseInfo{
					MessageId: "badId",
					NodeId:    n.id,
				})
			} else {
				c.responses.Push(ResponseInfo{
					MessageId: m.Id,
					NodeId:    n.id,
				})
			}

			messageList = append(messageList, m)
		}

		done := make(chan bool)
		errchan := make(chan error)
		go func() {
			for _, m := range messageList {
				err := c.Response(context.Background(), m.Id, m.Subject, m.Content)
				if err != nil {
					errchan <- err
					return
				}
			}

			done <- true
		}()

		select {
		case <-done:
		case err := <-errchan:
			if !tc.expectedFailure {
				t.Fatalf("could not send message: %s", err)
			}
			continue
		case <-time.After(time.Second * 3):
			t.Fatalf("did not Response in time")
		}

		if server.MessagesLength() != tc.nodeCount {
			t.Fatalf("expected %v messages to be sent but got %v instead", tc.nodeCount, server.MessagesLength())
		}
	}
}

/**************************************************************
Expected Outcomes:
- should return a channel
- should add a channel to the subject subscribed
**************************************************************/
func TestCourier_Subscribe(t *testing.T) {
	c, err := NewCourier(
		WithObserver(newMockObserver(make(chan []Noder), false)),
		WithHostname("test.com"),
		WithPort("3008"),
		WithDialOptions(grpc.WithInsecure()),
		WithFailedMessageWaitInterval(time.Second),
		WithMaxFailedMessageAttempts(5),
		StartOnCreation(false),
	)
	if err != nil {
		t.Fatalf("could not create new Courier: %s", err)
	}

	c.Subscribe("test")

	chanList, err := c.internalSubChannels.Subscriptions("test")
	if err != nil {
		t.Fatalf("couldn't get channels: %s", err)
	}

	if len(chanList) != 1 {
		t.Fatalf("expected length to be 1 but got %v", len(chanList))
	}
}

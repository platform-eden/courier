package messaging

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func TestNoObserverError_Error(t *testing.T) {
	method := "testMethod"
	e := &NoObserverError{
		Method: method,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: observer must be set", method) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: observer must be set", method), message)
	}
}

func TestSendCourierMessageError_Error(t *testing.T) {
	err := fmt.Errorf("test error")
	method := "testMethod"
	e := &SendCourierMessageError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

func TestCourierStartError_Error(t *testing.T) {
	err := fmt.Errorf("test error")
	method := "testMethod"
	e := &CourierStartError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

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
			dialOptions:     []grpc.DialOption{},
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
			WithClientNodeOptions(
				WithDialOptions(tc.dialOptions...),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			withCourierServer(testMessageServer),
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
		if len(c.clientNodeOptions) == 0 {
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
			WithClientNodeOptions(
				WithDialOptions(grpc.WithInsecure()),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
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

func TestCourier_Stop(t *testing.T) {
	type test struct {
		port     string
		observer Observer
	}

	tests := []test{
		{
			port:     "3000",
			observer: newMockObserver(make(chan []Noder), false),
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
			WithClientNodeOptions(
				WithDialOptions(grpc.WithInsecure()),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			withCourierServer(testMessageServer),
			StartOnCreation(true),
		)
		if err != nil {
			t.Fatalf("expected NewCourier to pass but it failed: %s", err)
		}

		c.Stop()

		if c.running == true {
			t.Fatal("expected runnning to be false but it's true")
		}
	}
}

/**************************************************************
Expected Outcomes:
- should return an error if the message passed doesn't have a topic registered by client nodes
- should send messages to all clients subscribed to the subject
**************************************************************/
func TestCourier_Publish(t *testing.T) {
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
		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			WithClientNodeOptions(
				WithInsecure(),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			withCourierServer(testMessageServer),
			StartOnCreation(true),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		for _, n := range nodes {
			c.clientSubscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
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

		done := make(chan struct{})
		errchan := make(chan error)
		go func() {
			err := c.Publish(context.Background(), tc.messageSubject, []byte("test"))
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
					c.Stop()
					close(errchan)
					t.Fatalf("could not request message: %s", err)
				}
				continue
			case <-time.After(time.Second * 3):
				t.Fatalf("did not send Publish in time")
			}
		}
		c.Stop()
		close(errchan)
	}
}

/**************************************************************
Expected Outcomes:
- should return an error if the message passed doesn't have a topic registered by client nodes
- should send messages to all clients subscribed to the subject
**************************************************************/
func TestCourier_Request(t *testing.T) {
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
		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			WithClientNodeOptions(
				WithDialOptions(grpc.WithInsecure()),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			withCourierServer(testMessageServer),
			StartOnCreation(true),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		for _, n := range nodes {
			c.clientSubscribers.Add(n.id, tc.nodeSubject)
			_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
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

		done := make(chan struct{})
		errchan := make(chan error)
		go func() {
			err := c.Request(context.Background(), tc.messageSubject, []byte("test"))
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
					c.Stop()
					close(errchan)
					t.Fatalf("could not request message: %s", err)
				}
				continue
			case <-time.After(time.Second * 3):
				t.Fatalf("did not send Response in time")
			}
		}
		c.Stop()
		close(errchan)
	}
}

/**************************************************************
Expected Outcomes:
- returns error if messageId isn't in responseMap
- sends a message to the node in the responseInfo
**************************************************************/
func TestCourier_Response(t *testing.T) {
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

		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Noder), false)),
			WithHostname("test.com"),
			withCourierServer(testMessageServer),
			WithClientNodeOptions(
				WithDialOptions(grpc.WithInsecure()),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			StartOnCreation(true),
		)
		if err != nil {
			t.Fatalf("could not create new Courier: %s", err)
		}

		messageIds := []string{}
		_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
		if err != nil {
			t.Fatalf("could not create mock client: %s", err)
		}

		for _, node := range nodes {
			clientNode, err := newClientNode(*node, c.Id, c.clientNodeOptions...)
			if err != nil {
				t.Fatalf("could not create client node: %s", err)
			}

			clientNode.connection = conn
			c.clientNodes.Add(*clientNode)

			id := uuid.NewString()
			messageIds = append(messageIds, id)

			if tc.badMessageId {
				c.responses.Push(ResponseInfo{
					NodeId:    node.id,
					MessageId: "badId",
				})
			} else {
				c.responses.Push(ResponseInfo{
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
				err := c.Response(ctx, id, tc.subject, []byte("test"))
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
					c.Stop()
					close(errchan)
					t.Fatalf("expected test to pass but it failed sending response: %s", err)
				}
			case <-time.After(time.Second * 10):
				t.Fatal("did not finish sending responses in time")
			}
		}

		c.Stop()
		close(errchan)
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
		WithClientNodeOptions(
			WithDialOptions(grpc.WithInsecure()),
			WithClientRetryOptions(ClientRetryOptionsInput{
				maxAttempts:     5,
				backOff:         time.Millisecond * 100,
				jitter:          0.2,
				perRetryTimeout: time.Second * 3,
			}),
		),
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

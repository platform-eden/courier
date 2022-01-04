package messaging

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestNoObserverError_Error(t *testing.T) {
	method := "testMethod"
	e := &NoObserverChannelError{
		Method: method,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: observer channel must be set", method) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: observer must be set", method), message)
	}
}

func TestSendMessagingServiceMessageError_Error(t *testing.T) {
	err := fmt.Errorf("test error")
	method := "testMethod"
	e := &SendMessagingServiceMessageError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

func TestMessagingServiceStartError_Error(t *testing.T) {
	err := fmt.Errorf("test error")
	method := "testMethod"
	e := &MessagingServiceStartError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

func TestNewMessagingService(t *testing.T) {
	type test struct {
		hostname        string
		port            string
		expectedFailure bool
		observerChannel chan []Noder
	}

	tests := []test{
		{
			hostname:        "test",
			port:            "3000",
			expectedFailure: false,
			observerChannel: make(chan []Noder),
		},
		{
			hostname:        "",
			port:            "3002",
			expectedFailure: false,
			observerChannel: make(chan []Noder),
		},
		{
			hostname:        "",
			port:            "3002",
			expectedFailure: true,
			observerChannel: nil,
		},
	}

	for i, tc := range tests {
		sub := []string{"sub1", "sub2", "sub3"}
		broad := []string{"broad1", "broad2", "broad3"}

		c, err := NewMessagingService(
			Subscribes(sub...),
			Broadcasts(broad...),
			WithHostname(tc.hostname),
			WithPort(tc.port),
			WithClientNodeOptions(
				WithInsecure(),
				WithClientRetryOptions(ClientRetryOptionsInput{
					maxAttempts:     5,
					backOff:         time.Millisecond * 100,
					jitter:          0.2,
					perRetryTimeout: time.Second * 3,
				}),
			),
			WithObserverChannel(tc.observerChannel),
			withMessagingServer(testMessageServer),
		)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("expected NewMessagingService to pass but it failed: %s", err)
		}

		if c.Hostname == "" {
			t.Fatal("Hostname should not be blank but it is")
		}
		if c.observerChannel == nil {
			t.Fatal("observerChannel should not be blank but it is")
		}
		if len(c.clientNodeOptions) == 0 {
			t.Fatal("DialOptions length should not be 0 but it is")
		}

		if tc.expectedFailure {
			t.Fatalf("%v: expected test to fail but it passed", i)
		}

		c.Stop()
	}
}

func TestMessagingService_Stop(t *testing.T) {
	type test struct {
		port string
	}

	tests := []test{
		{
			port: "3000",
		},
	}

	for _, tc := range tests {
		testMessageServer.SetToPass()
		sub := []string{"sub1", "sub2", "sub3"}
		broad := []string{"broad1", "broad2", "broad3"}
		ochan := make(chan []Noder)

		c, err := NewMessagingService(
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
			WithObserverChannel(ochan),
			withMessagingServer(testMessageServer),
		)
		if err != nil {
			t.Fatalf("expected NewMessagingService to pass but it failed: %s", err)
		}

		c.Stop()

		if c.running == true {
			t.Fatal("expected runnning to be false but it's true")
		}
	}
}

func TestMessagingService_Subscribe(t *testing.T) {
	ochan := make(chan []Noder)
	c, err := NewMessagingService(
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
		WithObserverChannel(ochan),
	)
	if err != nil {
		t.Fatalf("could not create new MessagingService: %s", err)
	}

	mockServer := NewMockServer(bufconn.Listen(1024*1024), "3000", false)
	c.server = mockServer
	c.Subscribe("test")

	chanList, err := mockServer.channelMap.Subscriptions("test")
	if err != nil {
		t.Fatalf("couldn't get channels: %s", err)
	}

	if len(chanList) != 1 {
		t.Fatalf("expected length to be 1 but got %v", len(chanList))
	}
}

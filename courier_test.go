package courier

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
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
		startOnCreation bool
		expectedFailure bool
	}

	tests := []test{
		{
			observer:        newMockObserver(make(chan []Node), false),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        nil,
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
			startOnCreation: false,
			expectedFailure: true,
		},
		{
			observer:        newMockObserver(make(chan []Node), false),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        newMockObserver(make(chan []Node), false),
			dialOptions:     []grpc.DialOption{},
			hostname:        "test",
			startOnCreation: true,
			expectedFailure: false,
		},
		{
			observer:        newMockObserver(make(chan []Node), true),
			dialOptions:     []grpc.DialOption{grpc.WithInsecure()},
			hostname:        "test",
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
			WithPort("3000"),
			WithClientContext(context.Background()),
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

		time.Sleep(time.Millisecond * 100)
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
			observer:        newMockObserver(make(chan []Node), false),
			expectedFailure: true,
		},
		{
			port:            "3000",
			observer:        newMockObserver(make(chan []Node), true),
			expectedFailure: true,
		},
		{
			port:            "3001",
			observer:        newMockObserver(make(chan []Node), false),
			expectedFailure: false,
		},
	}

	for _, tc := range tests {
		sub := []string{"sub1", "sub2", "sub3"}
		broad := []string{"broad1", "broad2", "broad3"}

		c, err := NewCourier(
			WithObserver(newMockObserver(make(chan []Node), false)),
			Subscribes(sub...),
			Broadcasts(broad...),
			WithHostname("test.com"),
			WithPort(tc.port),
			WithClientContext(context.Background()),
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

package client

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
// 	"github.com/platform-edn/courier/pkg/messaging"
// 	"github.com/platform-edn/courier/pkg/registry"
// 	"google.golang.org/grpc"
// )

// func TestClientNodeSendError_Error(t *testing.T) {
// 	err := fmt.Errorf("test error")
// 	method := "testMethod"
// 	e := &ClientNodeSendError{
// 		Method: method,
// 		Err:    err,
// 	}

// 	message := e.Error()

// 	if message != fmt.Sprintf("%s: %s", method, err) {
// 		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
// 	}
// }

// func TestClientNodeDialError_Error(t *testing.T) {
// 	hostname := "host"
// 	port := "8080"
// 	err := fmt.Errorf("test error")
// 	method := "testMethod"
// 	e := &ClientNodeDialError{
// 		Method:   method,
// 		Hostname: hostname,
// 		Port:     port,
// 		Err:      err,
// 	}

// 	message := e.Error()

// 	if message != fmt.Sprintf("%s: could not create connection at %s:%s: %s", method, hostname, port, err) {
// 		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: could not create connection at %s:%s: %s", method, hostname, port, err), message)
// 	}
// }

// func TestClientNodeMessageTypeError_Error(t *testing.T) {
// 	mType := messaging.PubMessage
// 	method := "testMethod"
// 	e := &ClientNodeMessageTypeError{
// 		Method: method,
// 		Type:   mType.String(),
// 	}

// 	message := e.Error()

// 	if message != fmt.Sprintf("%s: message must be of type %s", method, mType) {
// 		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: message must be of type %s", method, mType), message)
// 	}
// }

// /**************************************************************
// Expected Outcomes:
// - should fail if options are incorrect for creating a a grpc client connection
// **************************************************************/
// func TestNewClientNode(t *testing.T) {
// 	type test struct {
// 		port            string
// 		options         []ClientNodeOption
// 		expectedFailure bool
// 	}

// 	tests := []test{
// 		{
// 			options:         []ClientNodeOption{WithDialOptions(grpc.WithInsecure())},
// 			expectedFailure: false,
// 		},
// 		{
// 			options:         []ClientNodeOption{},
// 			expectedFailure: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		n := registry.CreateTestNodes(1, &registry.TestNodeOptions{})[0]
// 		n.Port = tc.port
// 		_, err := newClientNode(*n, uuid.NewString(), tc.options...)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("expected newClientNode to pass but it failed: %s", err)
// 		}

// 		if tc.expectedFailure {
// 			t.Fatal("expected newClientNode to fail but it passed")
// 		}
// 	}
// }

// /**************************************************************
// Expected Outcomes:
// - should send a publish message to a grpc server successfully
// - returns error if the message was not sent successfully
// - returns error if message is not a publish message
// **************************************************************/
// func TestClientNode_SendPublishMessage(t *testing.T) {
// 	defer testMessageServer.SetToPass()

// 	type test struct {
// 		m               Message
// 		serverFailure   bool
// 		expectedFailure bool
// 	}

// 	tests := []test{
// 		{
// 			m:               NewPubMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   true,
// 			expectedFailure: true,
// 		},
// 		{
// 			m:               NewPubMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: false,
// 		},
// 		{
// 			m:               NewReqMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		if tc.expectedFailure {
// 			testMessageServer.SetToFail()
// 		} else {
// 			testMessageServer.SetToPass()
// 		}

// 		opts := []grpc_retry.CallOption{
// 			grpc_retry.WithPerRetryTimeout(time.Second),
// 			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(time.Millisecond*100, 0.2)),
// 			grpc_retry.WithMax(3),
// 		}

// 		_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not creat client for server: %s", err)
// 		}
// 		defer conn.Close()

// 		cc := clientNode{
// 			connection: conn,
// 			currentId:  uuid.NewString(),
// 			Node:       *CreateTestNodes(1, &TestNodeOptions{})[0],
// 		}

// 		err = cc.sendPublishMessage(context.Background(), tc.m)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not send message: %s", err)
// 		}

// 		if tc.expectedFailure {
// 			t.Fatalf("sendPublishMessage was expected to fail but it didn't")
// 		}
// 	}
// }

// /**************************************************************
// Expected Outcomes:
// - should send a request message to a grpc server successfully
// - returns error if the message was not sent successfully
// - returns error if message is not a publish message
// **************************************************************/
// func TestClientNode_SendRequestMessage(t *testing.T) {
// 	defer testMessageServer.SetToPass()

// 	type test struct {
// 		m               Message
// 		serverFailure   bool
// 		expectedFailure bool
// 	}

// 	tests := []test{
// 		{
// 			m:               NewReqMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   true,
// 			expectedFailure: true,
// 		},
// 		{
// 			m:               NewReqMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: false,
// 		},
// 		{
// 			m:               NewPubMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		if tc.expectedFailure {
// 			testMessageServer.SetToFail()
// 		} else {
// 			testMessageServer.SetToPass()
// 		}

// 		_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not creat client for server: %s", err)
// 		}
// 		defer conn.Close()

// 		cc := clientNode{
// 			connection: conn,
// 			currentId:  uuid.NewString(),
// 			Node:       *CreateTestNodes(1, &TestNodeOptions{})[0],
// 		}

// 		err = cc.sendRequestMessage(context.Background(), tc.m)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not send message: %s", err)
// 		}

// 		if tc.expectedFailure {
// 			t.Fatalf("sendRequestMessage was expected to fail but it didn't")
// 		}
// 	}
// }

// /**************************************************************
// Expected Outcomes:
// - should send a request message to a grpc server successfully
// - returns error if the message was not sent successfully
// - returns error if message is not a publish message
// **************************************************************/
// func TestClientNode_SendResponseMessage(t *testing.T) {
// 	defer testMessageServer.SetToPass()

// 	type test struct {
// 		m               Message
// 		serverFailure   bool
// 		expectedFailure bool
// 	}

// 	tests := []test{
// 		{
// 			m:               NewRespMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   true,
// 			expectedFailure: true,
// 		},
// 		{
// 			m:               NewRespMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: false,
// 		},
// 		{
// 			m:               NewPubMessage(uuid.NewString(), "test", []byte("test")),
// 			serverFailure:   false,
// 			expectedFailure: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		if tc.expectedFailure {
// 			testMessageServer.SetToFail()
// 		} else {
// 			testMessageServer.SetToPass()
// 		}

// 		_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not creat client for server: %s", err)
// 		}
// 		defer conn.Close()

// 		cc := clientNode{
// 			connection: conn,
// 			currentId:  uuid.NewString(),
// 			Node:       *CreateTestNodes(1, &TestNodeOptions{})[0],
// 		}

// 		err = cc.sendResponseMessage(context.Background(), tc.m)
// 		if err != nil {
// 			if tc.expectedFailure {
// 				continue
// 			}
// 			t.Fatalf("could not send message: %s", err)
// 		}

// 		if tc.expectedFailure {
// 			t.Fatalf("sendResponseMessage was expected to fail but it didn't")
// 		}
// 	}
// }

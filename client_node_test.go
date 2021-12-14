package courier

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/google/uuid"
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
// 	mType := PubMessage
// 	method := "testMethod"
// 	e := &ClientNodeMessageTypeError{
// 		Method: method,
// 		Type:   mType,
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
// 		options         []grpc.DialOption
// 		expectedFailure bool
// 	}

// 	tests := []test{
// 		{
// 			options:         []grpc.DialOption{grpc.WithInsecure()},
// 			expectedFailure: false,
// 		},
// 		{
// 			options:         []grpc.DialOption{},
// 			expectedFailure: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		n := CreateTestNodes(1, &TestNodeOptions{})[0]
// 		n.port = tc.port
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

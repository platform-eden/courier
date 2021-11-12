package courier

import (
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

/**************************************************************
Expected Outcomes:
- should fail if options are incorrect for creating a a grpc client connection
**************************************************************/
func TestNewClientNode(t *testing.T) {
	type test struct {
		port            string
		options         []grpc.DialOption
		expectedFailure bool
	}

	tests := []test{
		{
			options:         []grpc.DialOption{grpc.WithInsecure()},
			expectedFailure: false,
		},
		{
			options:         []grpc.DialOption{},
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		n := CreateTestNodes(1, &TestNodeOptions{})[0]
		n.Port = tc.port
		_, err := newClientNode(*n, uuid.NewString(), tc.options...)
		if err != nil {
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("expected newClientNode to pass but it failed: %s", err)
		}

		if tc.expectedFailure {
			t.Fatal("expected newClientNode to fail but it passed")
		}
	}
}

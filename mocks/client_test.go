package mocks

import (
	"testing"

	"google.golang.org/grpc/test/bufconn"
)

func TestNewMockClient(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewMockServer(lis)

	_, conn, err := NewLocalGRPCClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not create client: %s", err)
	}

	defer conn.Close()
}

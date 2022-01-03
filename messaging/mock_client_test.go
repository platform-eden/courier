package messaging

import "testing"

func TestNewMockClient(t *testing.T) {
	_, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
	if err != nil {
		t.Fatalf("could not create client: %s", err)
	}

	defer conn.Close()
}

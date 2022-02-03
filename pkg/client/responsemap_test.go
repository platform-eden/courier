package client

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/pkg/messaging"
)

func TestUnregisteredResponseError_Error(t *testing.T) {
	method := "testMethod"
	messageId := "test"
	e := &UnregisteredResponseError{
		Method:    method,
		MessageId: messageId,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: no response exists with id %s", method, messageId) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: no response exists with id %s", method, messageId), message)
	}
}

func TestResponseMap_PushResponse(t *testing.T) {
	responses := newResponseMap()
	info := messaging.ResponseInfo{
		MessageId: "10",
		NodeId:    "node",
	}

	responses.PushResponse(info)
	id, ok := responses.responses[info.MessageId]
	if !ok {
		t.Fatal("expected response to be added but it wasn't")
	}
	if id != info.NodeId {
		t.Fatalf("expected response id to be %s but got %s", info.NodeId, id)
	}
}

func TestResponseMap_PopResponse(t *testing.T) {
	responses := newResponseMap()

	info := messaging.ResponseInfo{
		MessageId: "10",
		NodeId:    "node",
	}

	responses.PushResponse(info)
	id, err := responses.PopResponse(info.MessageId)
	if err != nil {
		t.Fatalf("error popping response: %s", err)
	}
	if id != info.NodeId {
		t.Fatalf("expected response id to be %s but got %s", info.NodeId, id)
	}

	_, err = responses.PopResponse(info.MessageId)
	if err == nil {
		t.Fatalf("expected popping a nonexistant response to return an err but it passed")
	}

}

/**************************************************************
Expected Outcomes:
- should return how many responseInfo structs are in the map
**************************************************************/
func TestResponseMap_Length(t *testing.T) {
	type test struct {
		count int
	}

	tests := []test{
		{
			count: 10,
		},
	}

	for _, tc := range tests {
		respMap := newResponseMap()
		for i := 0; i < tc.count; i++ {
			info := messaging.ResponseInfo{
				NodeId:    uuid.NewString(),
				MessageId: uuid.NewString(),
			}

			respMap.PushResponse(info)
		}

		if respMap.Length() != tc.count {
			t.Fatalf("expected response count to be %v but got %v", tc.count, respMap.Length())
		}
	}
}

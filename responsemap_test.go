package courier

import (
	"testing"

	"github.com/google/uuid"
)

func TestResponseMap_PushResponse(t *testing.T) {
	responses := newResponseMap()
	info := ResponseInfo{
		MessageId: "10",
		NodeId:    "node",
	}

	responses.Push(info)
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

	info := ResponseInfo{
		MessageId: "10",
		NodeId:    "node",
	}

	responses.Push(info)
	id, err := responses.Pop(info.MessageId)
	if err != nil {
		t.Fatalf("error popping response: %s", err)
	}
	if id != info.NodeId {
		t.Fatalf("expected response id to be %s but got %s", info.NodeId, id)
	}

	_, err = responses.Pop(info.MessageId)
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
			info := ResponseInfo{
				NodeId:    uuid.NewString(),
				MessageId: uuid.NewString(),
			}

			respMap.Push(info)
		}

		if respMap.Length() != tc.count {
			t.Fatalf("expected response count to be %v but got %v", tc.count, respMap.Length())
		}
	}
}

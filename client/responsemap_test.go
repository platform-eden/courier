package client

import (
	"testing"

	"github.com/platform-edn/courier/node"
)

func TestResponseMap_PushResponse(t *testing.T) {
	responses := newResponseMap()
	info := node.ResponseInfo{
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

	info := node.ResponseInfo{
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

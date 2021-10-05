package client

import (
	"testing"

	"github.com/platform-edn/courier/node"
)

func TestResponseMap_PushResponse(t *testing.T) {
	responses := newResponseMap()

	response := node.ResponseInfo{
		Id:      "10",
		Address: "test.com",
		Port:    "80",
	}

	responses.PushResponse(response)

	info, ok := responses.responses[response.Id]
	if !ok {
		t.Fatal("expected response to be added but it wasn't")
	}
	if info.Port != response.Port {
		t.Fatalf("expected response port to be %s but got %s", response.Port, info.Port)
	}
	if info.Address != response.Address {
		t.Fatalf("expected response address to be %s but got %s", response.Address, info.Address)
	}
}

func TestResponseMap_PopResponse(t *testing.T) {
	responses := newResponseMap()
	response := node.ResponseInfo{
		Id:      "10",
		Address: "test.com",
		Port:    "80",
	}

	responses.PushResponse(response)
	info, err := responses.PopResponse(response.Id)
	if err != nil {
		t.Fatalf("error popping response: %s", err)
	}
	if info.Port != response.Port {
		t.Fatalf("expected response port to be %s but got %s", response.Port, info.Port)
	}
	if info.Address != response.Address {
		t.Fatalf("expected response address to be %s but got %s", response.Address, info.Address)
	}

	_, err = responses.PopResponse(response.Id)
	if err == nil {
		t.Fatalf("expected popping a nonexistant response to return an err but it passed")
	}

}

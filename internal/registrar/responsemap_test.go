package registrar

import "testing"

func TestResponseMap_PushResponse(t *testing.T) {
	responses := NewResponseMap()
	id := "10"
	address := "test.com"
	port := "80"

	responses.PushResponse(id, address, port)

	info, ok := responses.responses[id]
	if !ok {
		t.Fatal("expected response to be added but it wasn't")
	}
	if info.Port != port {
		t.Fatalf("expected response port to be %s but got %s", port, info.Port)
	}
	if info.Address != address {
		t.Fatalf("expected response address to be %s but got %s", address, info.Address)
	}
}

func TestResponseMap_PopResponse(t *testing.T) {
	responses := NewResponseMap()
	id := "10"
	address := "test.com"
	port := "80"

	responses.PushResponse(id, address, port)
	info, err := responses.PopResponse(id)
	if err != nil {
		t.Fatalf("error popping response: %s", err)
	}
	if info.Port != port {
		t.Fatalf("expected response port to be %s but got %s", port, info.Port)
	}
	if info.Address != address {
		t.Fatalf("expected response address to be %s but got %s", address, info.Address)
	}

	_, err = responses.PopResponse(id)
	if err == nil {
		t.Fatalf("expected popping a nonexistant response to return an err but it passed")
	}

}

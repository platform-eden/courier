package client

import (
	"fmt"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/node"
)

type responseMap struct {
	responses map[string]*node.ResponseInfo
	lock      *lock.TicketLock
}

func newResponseMap() *responseMap {
	r := responseMap{
		responses: make(map[string]*node.ResponseInfo),
		lock:      lock.NewTicketLock(),
	}

	return &r
}

func (r *responseMap) PushResponse(response node.ResponseInfo) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.responses[response.Id] = &node.ResponseInfo{
		Address: response.Address,
		Port:    response.Port,
	}
}

func (r *responseMap) PopResponse(messageId string) (*node.ResponseInfo, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	info, ok := r.responses[messageId]
	if !ok {
		return nil, fmt.Errorf("response does not exist with %s as an id", messageId)
	}

	delete(r.responses, messageId)

	return info, nil
}

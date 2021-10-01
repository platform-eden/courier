package client

import (
	"fmt"

	"github.com/platform-edn/courier/lock"
)

type responseMap struct {
	responses map[string]*ResponseInfo
	lock      *lock.TicketLock
}

type ResponseInfo struct {
	Address string
	Port    string
}

func NewResponseMap() *responseMap {
	r := responseMap{
		responses: make(map[string]*ResponseInfo),
		lock:      lock.NewTicketLock(),
	}

	return &r
}

func (r *responseMap) PushResponse(messageId string, returnAdress string, returnPort string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.responses[messageId] = &ResponseInfo{
		Address: returnAdress,
		Port:    returnPort,
	}
}

func (r *responseMap) PopResponse(messageId string) (*ResponseInfo, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	info, ok := r.responses[messageId]
	if !ok {
		return nil, fmt.Errorf("response does not exist with %s as an id", messageId)
	}

	delete(r.responses, messageId)

	return info, nil
}

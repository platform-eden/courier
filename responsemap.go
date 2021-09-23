package courier

import "fmt"

type responseMap struct {
	responses map[string]*ResponseInfo
	lock      *ticketLock
}

type ResponseInfo struct {
	Address string
	Port    string
}

func NewResponseMap() *responseMap {
	r := responseMap{
		responses: make(map[string]*ResponseInfo),
		lock:      newTicketLock(),
	}

	return &r
}

func (r *responseMap) PushResponse(messageId string, returnAdress string, returnPort string) {
	r.lock.lock()
	defer r.lock.unlock()

	r.responses[messageId] = &ResponseInfo{
		Address: returnAdress,
		Port:    returnPort,
	}
}

func (r *responseMap) PopResponse(messageId string) (*ResponseInfo, error) {
	r.lock.lock()
	defer r.lock.unlock()

	info, ok := r.responses[messageId]
	if !ok {
		return nil, fmt.Errorf("response does not exist with %s as an id", messageId)
	}

	delete(r.responses, messageId)

	return info, nil
}

package registrar

import "fmt"

type ResponseMap struct {
	responses map[string]*ResponseInfo
	lock      *TicketLock
}

type ResponseInfo struct {
	Address string
	Port    string
}

func NewResponseMap() *ResponseMap {
	r := ResponseMap{
		responses: make(map[string]*ResponseInfo),
		lock:      newTicketLock(),
	}

	return &r
}

func (r *ResponseMap) PushResponse(messageId string, returnAdress string, returnPort string) {
	r.lock.lock()
	defer r.lock.unlock()

	r.responses[messageId] = &ResponseInfo{
		Address: returnAdress,
		Port:    returnPort,
	}
}

func (r *ResponseMap) PopResponse(messageId string) (*ResponseInfo, error) {
	r.lock.lock()
	defer r.lock.unlock()

	info, ok := r.responses[messageId]
	if !ok {
		return nil, fmt.Errorf("response does not exist with %s as an id", messageId)
	}

	delete(r.responses, messageId)

	return info, nil
}

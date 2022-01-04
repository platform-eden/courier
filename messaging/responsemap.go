package messaging

type responseMap struct {
	responses map[string]string
	lock      *TicketLock
}

func newResponseMap() *responseMap {
	r := responseMap{
		responses: make(map[string]string),
		lock:      NewTicketLock(),
	}

	return &r
}

func (r *responseMap) Push(info ResponseInfo) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.responses[info.MessageId] = info.NodeId
}

func (r *responseMap) Pop(messageId string) (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	nodeId, ok := r.responses[messageId]
	if !ok {
		return "", &UnregisteredResponseError{
			Method:    "Pop",
			MessageId: messageId,
		}
	}

	delete(r.responses, messageId)

	return nodeId, nil
}

func (r *responseMap) Length() int {
	r.lock.Lock()
	defer r.lock.Unlock()

	return len(r.responses)
}

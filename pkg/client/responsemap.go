package client

import (
	"fmt"

	"github.com/platform-edn/courier/pkg/lock"
	"github.com/platform-edn/courier/pkg/messaging"
)

type responseMap struct {
	responses map[string]string
	lock      *lock.TicketLock
}

func newResponseMap() *responseMap {
	r := responseMap{
		responses: make(map[string]string),
		lock:      lock.NewTicketLock(),
	}

	return &r
}

func (r *responseMap) PushResponse(info messaging.ResponseInfo) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.responses[info.MessageId] = info.NodeId
}

func (r *responseMap) PopResponse(messageId string) (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	nodeId, ok := r.responses[messageId]
	if !ok {
		return "", fmt.Errorf("PopResponse: %w", &UnregisteredResponseError{
			MessageId: messageId,
		})
	}

	delete(r.responses, messageId)

	return nodeId, nil
}

// generateIdsByMessage takes a messageId and returns a channel of node ids expecting to receive a response
func (r *responseMap) GenerateIdsByMessage(messageId string) (<-chan string, error) {
	out := make(chan string, 1)

	id, err := r.PopResponse(messageId)
	if err != nil {
		return nil, fmt.Errorf("GenerateIdsByMessage: %w", err)
	}

	out <- id
	close(out)

	return out, nil
}

func (r *responseMap) Length() int {
	r.lock.Lock()
	defer r.lock.Unlock()

	return len(r.responses)
}

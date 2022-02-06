package client

import (
	"fmt"

	"github.com/platform-edn/courier/pkg/lock"
	"github.com/platform-edn/courier/pkg/messaging"
)

type responseMap struct {
	Responses map[string]string
	lock.Locker
}

func NewResponseMap() *responseMap {
	r := responseMap{
		Responses: make(map[string]string),
		Locker:    lock.NewTicketLock(),
	}

	return &r
}

func (r *responseMap) PushResponse(info messaging.ResponseInfo) {
	r.Lock()
	defer r.Unlock()

	r.Responses[info.MessageId] = info.NodeId
}

func (r *responseMap) PopResponse(messageId string) (string, error) {
	r.Lock()
	defer r.Unlock()

	nodeId, ok := r.Responses[messageId]
	if !ok {
		return "", fmt.Errorf("PopResponse: %w", &UnregisteredResponseError{
			MessageId: messageId,
		})
	}

	delete(r.Responses, messageId)

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

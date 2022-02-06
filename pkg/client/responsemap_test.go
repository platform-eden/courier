package client_test

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

// func TestUnregisteredResponseError_Error(t *testing.T) {
// 	method := "testMethod"
// 	messageId := "test"
// 	e := &UnregisteredResponseError{
// 		Method:    method,
// 		MessageId: messageId,
// 	}

// 	message := e.Error()

// 	if message != fmt.Sprintf("%s: no response exists with id %s", method, messageId) {
// 		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: no response exists with id %s", method, messageId), message)
// 	}
// }

func TestResponseMap_PushResponse(t *testing.T) {
	tests := map[string]struct {
		responseInfo messaging.ResponseInfo
	}{
		"adds a pair where the key is the messageId": {
			responseInfo: messaging.ResponseInfo{
				MessageId: "message",
				NodeId:    "node",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			respMap := client.NewResponseMap()

			respMap.PushResponse(test.responseInfo)

			assert.Contains(respMap.Responses, test.responseInfo.MessageId)
			assert.Equal(respMap.Responses[test.responseInfo.MessageId], test.responseInfo.NodeId)
		})
	}
}

func TestResponseMap_PopResponse(t *testing.T) {
	tests := map[string]struct {
		responseInfo messaging.ResponseInfo
		popId        string
		err          error
	}{
		"pops a response from the map": {
			responseInfo: messaging.ResponseInfo{
				MessageId: "message",
				NodeId:    "node",
			},
			popId: "message",
			err:   nil,
		},
		"fails when popping a message that isn't registered": {
			responseInfo: messaging.ResponseInfo{
				MessageId: "message",
				NodeId:    "node",
			},
			popId: "bad",
			err:   &client.UnregisteredResponseError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			respMap := client.NewResponseMap()

			respMap.PushResponse(test.responseInfo)
			id, err := respMap.PopResponse(test.popId)
			if test.err != nil {
				errorType := test.err
				assert.Error(err)
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.Equal(id, test.responseInfo.NodeId)
			assert.Zero(len(respMap.Responses))
		})
	}
}

func TestResponseMap_GenerateIdsByMessage(t *testing.T) {
	tests := map[string]struct {
		responseInfo messaging.ResponseInfo
		id           string
		err          error
	}{
		"pops a response from the map": {
			responseInfo: messaging.ResponseInfo{
				MessageId: "message1",
				NodeId:    "node",
			},
			id:  "message1",
			err: nil,
		},
		"fails when popping a message that isn't registered": {
			responseInfo: messaging.ResponseInfo{
				MessageId: "message1",
				NodeId:    "node",
			},
			id:  "bad1",
			err: &client.UnregisteredResponseError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			respMap := client.NewResponseMap()

			respMap.PushResponse(test.responseInfo)

			out, err := respMap.GenerateIdsByMessage(test.id)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)

			done := make(chan struct{})
			go func() {
				respOut := []string{}
				for s := range out {
					respOut = append(respOut, s)
				}

				assert.Len(respOut, 1)
				assert.EqualValues(respOut[0], test.responseInfo.NodeId)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second * 3):
				t.Fatal("did not close done channel in time")
			}
		})
	}
}

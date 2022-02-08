package server_test

import (
	"fmt"
	"testing"

	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestChannelMap_Add(t *testing.T) {
	type test struct {
		channelCount int
		subject      string
	}

	tests := []test{
		{
			channelCount: 10,
			subject:      "test",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			assert := assert.New(t)
			chanMap := server.NewChannelMap()
			defer chanMap.CloseSubscriberChannels()

			for i := 0; i < tc.channelCount; i++ {
				chanMap.SubscribeToSubject(tc.subject)
			}

			assert.Len(chanMap.SubjectChannels[tc.subject], tc.channelCount)
		})
	}
}

func TestChannelMap_Subscriptions(t *testing.T) {
	tests := map[string]struct {
		addSubject          string
		subscriptionSubject string
		channelCount        int
		err                 error
	}{
		"all subscribers to a subject is returned": {
			addSubject:          "test",
			subscriptionSubject: "test",
			channelCount:        10,
			err:                 nil,
		},
		"no subject error is returned": {
			addSubject:          "test",
			subscriptionSubject: "test1",
			channelCount:        10,
			err:                 &server.UnregisteredChannelSubjectError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			chanMap := server.NewChannelMap()
			defer chanMap.CloseSubscriberChannels()

			for i := 0; i < test.channelCount; i++ {
				chanMap.SubscribeToSubject(test.addSubject)
			}

			channels, err := chanMap.Subscriptions(test.subscriptionSubject)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.Len(channels, test.channelCount)
		})
	}
}

func TestChannelMap_GenerateMessageChannels(t *testing.T) {
	tests := map[string]struct {
		addSubject          string
		subscriptionSubject string
		channelCount        int
		err                 error
	}{
		"all subscribers to a subject is returned": {
			addSubject:          "test",
			subscriptionSubject: "test",
			channelCount:        10,
			err:                 nil,
		},
		"no subject error is returned": {
			addSubject:          "test",
			subscriptionSubject: "test1",
			channelCount:        10,
			err:                 &server.UnregisteredChannelSubjectError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			chanMap := server.NewChannelMap()
			defer chanMap.CloseSubscriberChannels()

			for i := 0; i < test.channelCount; i++ {
				chanMap.SubscribeToSubject(test.addSubject)
			}

			channels, err := chanMap.GenerateMessageChannels(test.subscriptionSubject)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)

			chanList := []chan messaging.Message{}
			for channel := range channels {
				chanList = append(chanList, channel)
			}

			assert.Len(chanList, test.channelCount)
		})
	}
}

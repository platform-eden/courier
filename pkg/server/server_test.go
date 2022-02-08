package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/server"
)

func TestMessageServer_PublishMessage(t *testing.T) {
	type test struct {
		clientSubject   string
		chanMapSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			clientSubject:   "test",
			chanMapSubject:  "test",
			expectedFailure: false,
		},
		{
			clientSubject:   "test",
			chanMapSubject:  "test1",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		rchan := make(chan messaging.ResponseInfo)
		defer close(rchan)
		errchan := make(chan error)
		ms := server.NewMessagingServer()

		mchan := ms.ChannelMapper.SubscribeToSubject(tc.chanMapSubject)

		go func() {
			m := messaging.PublishMessageRequest{
				Message: &messaging.PublishMessage{
					Id:      uuid.NewString(),
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			}

			_, err := ms.PublishMessage(context.Background(), &m)
			if err != nil {
				errchan <- err
			}
		}()

		select {
		case <-mchan:
		case err := <-errchan:
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("failed serving PublishMessage: %s", err)
		case <-time.After(time.Second):
			t.Fatal("didn't receive message on Message Channel back in time")
		}

		ms.ChannelMapper.CloseSubscriberChannels()
	}
}

func TestMessageServer_ResponseMessage(t *testing.T) {
	type test struct {
		clientSubject   string
		chanMapSubject  string
		expectedFailure bool
	}

	tests := []test{
		{
			clientSubject:   "test",
			chanMapSubject:  "test",
			expectedFailure: false,
		},
		{
			clientSubject:   "test",
			chanMapSubject:  "test1",
			expectedFailure: true,
		},
	}

	for _, tc := range tests {
		rchan := make(chan messaging.ResponseInfo)
		defer close(rchan)
		errchan := make(chan error)
		ms := server.NewMessagingServer()

		mchan := ms.ChannelMapper.SubscribeToSubject(tc.chanMapSubject)

		go func() {
			m := messaging.ResponseMessageRequest{
				Message: &messaging.ResponseMessage{
					Id:      uuid.NewString(),
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			}

			_, err := ms.ResponseMessage(context.Background(), &m)
			if err != nil {
				errchan <- err
			}
		}()

		select {
		case <-mchan:
		case err := <-errchan:
			if tc.expectedFailure {
				continue
			}
			t.Fatalf("failed serving ResponseMessage: %s", err)
		case <-time.After(time.Second):
			t.Fatal("didn't receive message on Message Channel back in time")
		}

		ms.ChannelMapper.CloseSubscriberChannels()
	}
}

// func Test(t *testing.T) {
// 	tests := map[string]struct {
// 		clientSubject  string
// 		chanMapSubject string
// 		err            error
// 	}{}

// 	for name, test := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			assert := assert.New(t)
// 			rchan := make(chan messaging.ResponseInfo, 1)
// 			ms := server.NewMessagingServer()

// 			mchan := ms.ChannelMapper.SubscribeToSubject(test.chanMapSubject)
// 			errchan := make(chan error)

// 			go func() {
// 				_, err := ms.RequestMessage(context.Background(), &messaging.RequestMessageRequest{
// 					Message: &messaging.RequestMessage{
// 						Id:      "messageId",
// 						NodeId:  "nodeId",
// 						Subject: test.clientSubject,
// 						Content: []byte("test"),
// 					},
// 				})
// 				if err != nil {
// 					errchan <- err
// 				}
// 			}()

// 			select {
// 			case <-mchan:
// 			case err := <-errchan:
// 				if test.expectedFailure {
// 					continue
// 				}
// 				t.Fatalf("failed serving RequestMessage: %s", err)
// 			case <-time.After(time.Second):
// 				t.Fatal("didn't receive message on Message Channel back in time")
// 			}

// 			close(rchan)

// 			count := 0
// 			for range rchan {
// 				count++
// 			}

// 			if count != 1 {
// 				t.Fatalf("expected a ResponseInfo to be sent through but got %v ResponseInfos", count)
// 			}

// 			ms.ChannelMapper.CloseSubscriberChannels()
// 		})
// 	}
// }

// func TestServer_FanForwardMessages(t *testing.T) {
// 	type test struct {
// 		channelCount int
// 	}

// 	tests := []test{
// 		{
// 			channelCount: 10,
// 		},
// 	}

// 	for _, tc := range tests {
// 		ms := server.NewMessagingServer()
// 		mchan := make(chan messaging.Message, tc.channelCount)
// 		cchan := make(chan chan messaging.Message)
// 		m := messaging.NewPubMessage("test", "test", []byte("test"))

// 		forward := func(mchan chan messaging.Message, m messaging.Message, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			mchan <- m
// 		}

// 		go func() {
// 			for i := 0; i < tc.channelCount; i++ {
// 				cchan <- mchan
// 			}
// 			close(cchan)
// 		}()

// 		ms.FanForwardMessages(cchan, m, forward)
// 		close(mchan)

// 		count := 0
// 		for range mchan {
// 			count++
// 		}

// 		if count != tc.channelCount {
// 			t.Fatalf("expected passed messages to be %v but got %v", tc.channelCount, count)
// 		}
// 	}
// }

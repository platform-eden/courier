package messaging

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/messaging/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMessageServerStartError_Error(t *testing.T) {
	method := "testMethod"
	err := fmt.Errorf("test error")
	e := &MessageServerStartError{
		Method: method,
		Err:    err,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: %s", method, err) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: %s", method, err), message)
	}
}

/**************************************************************
Expected Outcomes:
- all subject subscribers should receive a message that is assigned that subject
- should return an error to the client is subject doesn't exist in channelMap
**************************************************************/
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
		rchan := make(chan ResponseInfo)
		defer close(rchan)
		chanMap := newChannelMap()
		defer chanMap.Close()
		mchan := chanMap.Add(tc.chanMapSubject)
		errchan := make(chan error)
		ms := NewMessageServer("3000", rchan, chanMap)

		go func() {
			m := proto.PublishMessageRequest{
				Message: &proto.PublishMessage{
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
	}
}

/**************************************************************
Expected Outcomes:
- all subject subscribers should receive a message that is assigned that subject
- should return an error to the client is subject doesn't exist in channelMap
**************************************************************/
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
		rchan := make(chan ResponseInfo)
		defer close(rchan)
		chanMap := newChannelMap()
		defer chanMap.Close()
		mchan := chanMap.Add(tc.chanMapSubject)
		errchan := make(chan error)
		ms := NewMessageServer("3000", rchan, chanMap)

		go func() {
			m := proto.ResponseMessageRequest{
				Message: &proto.ResponseMessage{
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
	}
}

/**************************************************************
Expected Outcomes:
- all subject subscribers should receive a message that is assigned that subject
- should return an error to the client is subject doesn't exist in channelMap
- should send a resposeInfo through rchan
**************************************************************/
func TestMessageServer_RequestMessage(t *testing.T) {
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
		rchan := make(chan ResponseInfo, 1)
		chanMap := newChannelMap()
		ms := NewMessageServer("3000", rchan, chanMap)
		mchan := chanMap.Add(tc.chanMapSubject)
		defer chanMap.Close()
		errchan := make(chan error)

		go func() {
			_, err := ms.RequestMessage(context.Background(), &proto.RequestMessageRequest{
				Message: &proto.RequestMessage{
					Id:      "messageId",
					NodeId:  "nodeId",
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			})
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
			t.Fatalf("failed serving RequestMessage: %s", err)
		case <-time.After(time.Second):
			t.Fatal("didn't receive message on Message Channel back in time")
		}

		close(rchan)

		count := 0
		for range rchan {
			count++
		}

		if count != 1 {
			t.Fatalf("expected a ResponseInfo to be sent through but got %v ResponseInfos", count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- every channel in the passed in list should be passed through the out channel
**************************************************************/
func TestGenerateMessageChannels(t *testing.T) {
	type test struct {
		channelCount int
	}

	tests := []test{
		{
			channelCount: 10,
		},
	}

	for _, tc := range tests {
		channels := []chan Message{}

		for i := 0; i < tc.channelCount; i++ {
			channel := make(chan Message)

			channels = append(channels, channel)
		}

		out := generateMessageChannels(channels)

		count := 0
		for range out {
			count++
		}

		if count != tc.channelCount {
			t.Fatalf("expected channel count to be %v but got %v", tc.channelCount, count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- forward function is called for every channel passed in
- doesn't complete until all forward function completes
**************************************************************/
func TestFanForwardMessages(t *testing.T) {
	type test struct {
		channelCount int
	}

	tests := []test{
		{
			channelCount: 10,
		},
	}

	for _, tc := range tests {
		mchan := make(chan Message, tc.channelCount)
		cchan := make(chan chan Message)
		m := NewPubMessage("test", "test", []byte("test"))

		forward := func(mchan chan Message, m Message, wg *sync.WaitGroup) {
			defer wg.Done()
			mchan <- m
		}

		go func() {
			for i := 0; i < tc.channelCount; i++ {
				cchan <- mchan
			}
			close(cchan)
		}()

		fanForwardMessages(cchan, m, forward)
		close(mchan)

		count := 0
		for range mchan {
			count++
		}

		if count != tc.channelCount {
			t.Fatalf("expected passed messages to be %v but got %v", tc.channelCount, count)
		}
	}
}

/**************************************************************
Expected Outcomes:
- message passed into the function should be sent through the channel passed in
**************************************************************/
func TestForwardMessage(t *testing.T) {
	mchan := make(chan Message, 1)
	m := NewPubMessage("test", "test", []byte("test"))
	wg := &sync.WaitGroup{}

	wg.Add(1)
	forwardMessage(mchan, m, wg)
	wg.Wait()
	close(mchan)

	count := 0
	for range mchan {
		count++
	}

	if count != 1 {
		t.Fatalf("expected count to equal 1 but got %v", count)
	}
}

var lis = bufconn.Listen(1024 * 1024)

/**************************************************************
Expected Outcomes:
- should start a grpc server
**************************************************************/
func TestStartCourierServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	server := grpc.NewServer()

	go startCourierServer(ctx, &wg, server, lis, "3005")

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 3):
		t.Fatal("wait group did not finish in time")
	}
}

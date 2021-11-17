package courier

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
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
		server := NewMessageServer(rchan, chanMap)
		errchan := make(chan error)

		go startTestServer(errchan, server)

		client, conn, err := NewLocalGRPCClient("bufnet2", bufDialer)
		if err != nil {
			t.Fatalf("could not create client: %s", err)
		}
		defer conn.Close()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
			defer cancel()
			m := proto.PublishMessageRequest{
				Message: &proto.PublishMessage{
					Id:      uuid.NewString(),
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			}

			_, err := client.PublishMessage(ctx, &m)
			if err != nil {
				errchan <- err
			}
		}()

		if tc.expectedFailure {
			select {
			case <-errchan:
				continue
			case <-time.After(time.Second * 3):
				t.Fatal("PublishMessage didn't send err message in time")
			}
		}

		select {
		case <-mchan:
		case err := <-errchan:
			t.Fatalf("expected PublishMessage to pass but it failed: %s", err)
		case <-time.After(time.Second * 3):
			t.Fatal("PublishMessage didn't send message in time")
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
		server := NewMessageServer(rchan, chanMap)
		errchan := make(chan error)

		go startTestServer(errchan, server)

		client, conn, err := NewLocalGRPCClient("bufnet1", bufDialer)
		if err != nil {
			t.Fatalf("could not create client: %s", err)
		}
		defer conn.Close()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
			defer cancel()
			m := proto.ResponseMessageRequest{
				Message: &proto.ResponseMessage{
					Id:      uuid.NewString(),
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			}

			_, err := client.ResponseMessage(ctx, &m)
			if err != nil {
				errchan <- err
			}
		}()

		if tc.expectedFailure {
			select {
			case <-errchan:
				continue
			case <-time.After(time.Second * 3):
				t.Fatal("ResponseMessage didn't send err message in time")
			}
		}

		select {
		case <-mchan:
		case err := <-errchan:
			t.Fatalf("expected ResponseMessage to pass but it failed: %s", err)
		case <-time.After(time.Second * 3):
			t.Fatal("ResponseMessage didn't send message in time")
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
		defer chanMap.Close()
		mchan := chanMap.Add(tc.chanMapSubject)
		server := NewMessageServer(rchan, chanMap)
		errchan := make(chan error)

		go startTestServer(errchan, server)

		client, conn, err := NewLocalGRPCClient("buf", bufDialer)
		if err != nil {
			t.Fatalf("could not create client: %s", err)
		}
		defer conn.Close()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
			defer cancel()
			m := proto.RequestMessageRequest{
				Message: &proto.RequestMessage{
					Id:      uuid.NewString(),
					Subject: tc.clientSubject,
					Content: []byte("test"),
				},
			}

			_, err := client.RequestMessage(ctx, &m)
			if err != nil {
				errchan <- err
			}
		}()

		if tc.expectedFailure {
			select {
			case <-errchan:
				continue
			case <-time.After(time.Second * 3):
				t.Fatal("RequestMessage didn't send err message in time")
			}
		}

		select {
		case <-mchan:
		case err := <-errchan:
			t.Fatalf("expected RequestMessage to pass but it failed: %s", err)
		case <-time.After(time.Second * 3):
			t.Fatal("RequestMessage didn't send message in time")
		}

		close(rchan)

		count := 0
		for range rchan {
			count++
		}

		if count != 1 {
			t.Fatalf("expected count to be 1 but got %v", count)
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

func startTestServer(errchan chan error, m *MessageServer) {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()

	proto.RegisterMessageServerServer(grpcServer, m)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errchan <- fmt.Errorf("Server exited with error: %v", err)
		}
	}()
}

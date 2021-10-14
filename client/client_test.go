package client

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/mock"
	"github.com/platform-edn/courier/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMessageClient_PushChannel(t *testing.T) {
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	ctx := context.Background()
	address := "test.com"
	port := "80"

	client := NewMessageClient(ichan, nchan, fcchan, ctx, address, port, []grpc.DialOption{})
	pchan := client.pushChannel

	if client.pushChannel != pchan {
		t.Fatalf("expected push channel to be the same as one given but it wasn't")
	}
}

func TestMessageClient_ListenForResponseInfo(t *testing.T) {
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	ctx := context.Background()
	address := "test.com"
	port := "80"
	client := NewMessageClient(ichan, nchan, fcchan, ctx, address, port, []grpc.DialOption{})
	id := uuid.NewString()

	r := node.ResponseInfo{
		Id:      id,
		Address: "test.com",
		Port:    "80",
	}

	ichan <- r
	pass := make(chan bool)

	go func() {
		for {
			_, err := client.responseMap.PopResponse(id)
			if err != nil {
				time.Sleep(time.Second / 4)
				continue
			}
			pass <- true
			break
		}
	}()

	select {
	case <-pass:
		break
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on response channel")
	}
}

func TestMessageClient_ListenForSubscribers(t *testing.T) {
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	ctx := context.Background()
	address := "test.com"
	port := "80"
	client := NewMessageClient(ichan, nchan, fcchan, ctx, address, port, []grpc.DialOption{})

	nodecount := 10
	subjects := []string{"sub1"}
	nodes := mock.CreateTestNodes(nodecount, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})

	nm := map[string]node.Node{}

	for _, n := range nodes {
		nm[n.Id] = *n
	}

	nchan <- nm
	pass := make(chan bool)
	err := make(chan error)
	defer close(pass)
	defer close(err)

	go func() {
		for {
			if client.subscriberMap.TotalSubscribers("sub1") == 0 {
				continue
			}
			subs, e := client.subscriberMap.SubjectSubscribers("sub1")
			if e != nil {
				err <- e
			}
			if len(subs) != nodecount {
				time.Sleep(time.Second / 4)
				continue
			}
			pass <- true
			break
		}
	}()

	select {
	case <-pass:
		break
	case e := <-err:
		t.Fatal(e)
	case <-time.After(time.Second * 1):
		t.Fatal("client's subscribermap took too long to update")
	}
}

func TestMessageClient_ListenForOutgoingMessages(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := mock.NewMockServer(lis)
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	ctx := context.Background()
	address := "test.com"
	port := "80"

	subjects := []string{"test"}
	n := mock.CreateTestNodes(1, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]
	n.Address = "bufnet"
	n.Port = ""
	options := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(s.BufDialer),
	}
	client := NewMessageClient(ichan, nchan, fcchan, ctx, address, port, options)
	pchan := client.pushChannel

	client.subscriberMap.AddSubscriber(*n)

	m := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	m1 := message.NewReqMessage(uuid.NewString(), "test", []byte("test"))

	rid := uuid.NewString()
	m2 := message.NewRespMessage(rid, "test", []byte("test"))

	client.responseMap.PushResponse(node.ResponseInfo{
		Id:      rid,
		Address: "bufnet",
		Port:    "",
	})

	pchan <- m
	pchan <- m1
	pchan <- m2

	pass := make(chan bool)
	defer close(pass)
	go func() {
		for {
			if s.MessagesLength() == 3 {
				pass <- true
				break
			}
			time.Sleep(time.Second / 4)
		}
	}()

	select {
	case <-pass:
		break
	case <-time.After(time.Second * 1):
		t.Fatal("took to long to send messages")
	}
}

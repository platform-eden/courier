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
)

func TestMessageClient_PushChannel(t *testing.T) {
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	ctx := context.Background()
	id := uuid.NewString()

	client := NewMessageClient(
		id,
		ichan,
		nchan,
		fcchan,
		WithContext(ctx),
	)
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
	id := uuid.NewString()

	client := NewMessageClient(
		id,
		ichan,
		nchan,
		fcchan,
		WithContext(ctx),
	)
	mid := uuid.NewString()
	nid := uuid.NewString()

	r := node.ResponseInfo{
		MessageId: mid,
		NodeId:    nid,
	}

	ichan <- r
	pass := make(chan bool)

	go func() {
		for {
			_, err := client.responseMap.PopResponse(mid)
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
	id := uuid.NewString()

	client := NewMessageClient(
		id,
		ichan,
		nchan,
		fcchan,
		WithContext(ctx),
	)

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
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	id := uuid.NewString()
	ctx := context.Background()
	pubs := mock.NewMockSender()
	reqs := mock.NewMockSender()
	resps := mock.NewMockSender()

	c := NewMessageClient(
		id,
		ichan,
		nchan,
		fcchan,
		WithContext(ctx),
		WithDialOption(grpc.WithInsecure()),
		WithFailedWaitInterval(time.Millisecond*300),
		WithMaxFailedAttempts(3),
		WithPublishSender(pubs),
		WithRequestSender(reqs),
		WithResponseSender(resps),
	)

	subjects := []string{"test"}
	n := mock.CreateTestNodes(1, &mock.TestNodeOptions{
		SubscribedSubjects:  subjects,
		BroadcastedSubjects: []string{"broad1", "broad2", "broad3"},
	})[0]

	r := node.ResponseInfo{
		MessageId: uuid.NewString(),
		NodeId:    n.Id,
	}

	c.subscriberMap.AddSubscriber(*n)
	c.responseMap.PushResponse(r)

	push := c.PushChannel()

	pubm := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	push <- pubm
	reqm := message.NewReqMessage(uuid.NewString(), "test", []byte("test"))
	push <- reqm
	respm := message.NewRespMessage(r.MessageId, "test", []byte("test"))
	push <- respm

	time.Sleep(time.Second)

	if pubs.Count(mock.Publish) != 1 {
		t.Fatal("expected to have sent publish message but it didn't")
	}
	if reqs.Count(mock.Request) != 1 {
		t.Fatal("expected to have sent request message but it didn't")
	}
	if resps.Count(mock.Response) != 1 {
		t.Fatal("expected to have sent response message but it didn't")
	}

}

func TestMessageClient_AttemptMessage(t *testing.T) {
	ichan := make(chan node.ResponseInfo)
	defer close(ichan)
	nchan := make(chan map[string]node.Node)
	defer close(nchan)
	fcchan := make(chan node.Node)
	defer close(fcchan)
	id := uuid.NewString()
	ctx := context.Background()

	c := NewMessageClient(
		id,
		ichan,
		nchan,
		fcchan,
		WithContext(ctx),
		WithDialOption(grpc.WithInsecure()),
		WithFailedWaitInterval(time.Millisecond*300),
		WithMaxFailedAttempts(3))
	n := mock.CreateTestNodes(1, &mock.TestNodeOptions{})[0]
	m := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	s := mock.NewMockSender()

	go c.attemptMessage(m, n, s)

passloop:
	for {
		cycles := 0
		select {
		case <-c.failedConnChannel:
			t.Fatalf("error attempting message when it should not have")
		default:
			if s.Length() != 1 {
				if cycles > c.maxFailedAttempts {
					t.Fatalf("message was never sent on sender")
				}
				cycles++
				time.Sleep(c.failedWaitInterval)
				continue
			}

			break passloop
		}
	}

	s.Fail = true
	go c.attemptMessage(m, n, s)

failloop:
	for {
		select {
		case <-c.failedConnChannel:
			break failloop
		case <-time.After(time.Second):
			t.Fatalf("never received failed node in expected time")
		}
	}
}

func TestCreateGRPCClient(t *testing.T) {
	address := "test.com"
	port := "8080"
	options := []grpc.DialOption{grpc.WithInsecure()}

	_, _, err := createGRPCClient(address, port, options...)
	if err != nil {
		t.Fatalf("error creating grpc client: %s", err)
	}
}

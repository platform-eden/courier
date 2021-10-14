package courier

import (
	"context"
	"testing"
	"time"

	"github.com/platform-edn/courier/mock"
	"google.golang.org/grpc"
)

func TestNewCourier(t *testing.T) {
	sub := []string{"sub1", "sub2", "sub3"}
	broad := []string{"broad1", "broad2", "broad3"}

	nodes := mock.CreateTestNodes(5, &mock.TestNodeOptions{
		SubscribedSubjects:  sub,
		BroadcastedSubjects: broad,
	})

	store := mock.NewMockNodeStore(nodes...)
	NewCourier(store,
		WithId("test"),
		Subscribes(sub...),
		Broadcasts(broad...),
		ListensOnAddress("test.com"),
		ListensOnPort("3000"),
		WithClientContext(context.Background()),
		WithDialOption(grpc.WithInsecure()),
		WithFailedMessageWaitInterval(time.Second),
		WithDialOption(grpc.WithInsecure()),
		WithMaxFailedMessageAttempts(5),
		WithNodeStoreInterval(time.Second),
	)
}

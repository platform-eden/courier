package courier

import (
	"context"
	"testing"

	"github.com/platform-edn/courier/mock"
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
		WithClientMessageContext(context.Background()),
	)
}

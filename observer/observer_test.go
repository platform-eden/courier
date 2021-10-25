package observer

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/mock"
	"github.com/platform-edn/courier/node"
)

func TestStoreObserver_Observe(t *testing.T) {
	observer := mock.NewMockObserver(time.Second * 1)

	go observe(observer)

	observer.FailedConnectionChannel() <- node.Node{}

	if observer.BlackListCount() != 1 {
		t.Fatalf("expected blacklist count to equal 1 but got %v", observer.BlackListCount())
	}
}

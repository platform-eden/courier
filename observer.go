package courier

import "time"

type storeObserver struct {
	Store           NodeStorer
	ObserveInterval time.Duration
	CurrentNodes    []*node
	Subscribers     *subscriberMap
}

func NewStoreObserver(store NodeStorer, interval time.Duration) *storeObserver {
	s := storeObserver{
		Store:           store,
		ObserveInterval: interval,
		CurrentNodes:    []*node{},
	}

	return &s
}

func (s *storeObserver) start() {

}

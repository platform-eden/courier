package courier

import (
	"errors"
	"time"

	"github.com/platform-edn/courier/observer"
)

type CourierOptions struct {
	NodeStore           observer.NodeStorer
	ObserveInterval     time.Duration
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Address             string
	Port                string
}

type Courier struct {
	storeObserver       *observer.StoreObserver
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Address             string
	Port                string
}

func NewCourier(options CourierOptions) (*Courier, error) {
	if options.NodeStore == nil {
		return nil, errors.New("must have a NodeStore set in order to instantiate courier service")
	}
	if options.Address == "" {
		return nil, errors.New("must have an Address set in order to instantiate courier service")
	}
	if options.Port == "" {
		return nil, errors.New("must have a Port set in order to instantiate courier service")
	}

	c := Courier{
		BroadcastedSubjects: options.BroadcastedSubjects,
		SubscribedSubjects:  options.SubscribedSubjects,
		Address:             options.Address,
		Port:                options.Port,
	}

	if len(options.BroadcastedSubjects) != 0 {
		if options.ObserveInterval == 0 {
			return nil, errors.New("cannot have a time interval that is unset or equal to 0")
		}

		c.storeObserver = observer.NewStoreObserver(options.NodeStore, options.ObserveInterval, options.BroadcastedSubjects)
		// clientChannel := c.storeObserver.listenChannel()
		c.storeObserver.Start()
	}

	return &c, nil
}

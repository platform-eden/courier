package observe

import (
	"errors"
	"time"

	"github.com/platform-edn/courier/node"
)

type ObserveOption func(c *ObserveOptions)

func WithNodeStorer(store NodeStorer) ObserveOption {
	return func(o *ObserveOptions) {
		o.Store = store
	}
}

func WithInterval(interval time.Duration) ObserveOption {
	return func(o *ObserveOptions) {
		o.Interval = interval
	}
}

func WithSubjects(subjects []string) ObserveOption {
	return func(o *ObserveOptions) {
		o.Subjects = subjects
	}
}

type ObserveOptions struct {
	Store                   NodeStorer
	Interval                time.Duration
	Subjects                []string
	FailedConnectionChannel chan node.Node
	NewNodeChannel          chan node.Node
}

func NewObserveOptions(options ...ObserveOption) (*ObserveOptions, error) {
	o := &ObserveOptions{
		Store:                   nil,
		Interval:                time.Second,
		Subjects:                []string{},
		FailedConnectionChannel: make(chan node.Node),
		NewNodeChannel:          make(chan node.Node),
	}

	for _, option := range options {
		option(o)
	}

	if o.Store == nil {
		return nil, errors.New("store cannot be empty in ObserveOptions")
	}

	return o, nil
}

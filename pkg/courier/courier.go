package courier

import (
	"context"
	"fmt"
	"sync"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/eventer"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/platform-edn/courier/pkg/server"
)

type NodeRegister interface {
	EventInChannel() chan registry.NodeEvent
	SubscribeToEvents() chan registry.NodeEvent
	RegisterNodes(context.Context, *sync.WaitGroup, chan error)
}

type NodeEventer interface {
	DiscoverNodeEvents(context.Context, chan registry.NodeEvent, chan error, *sync.WaitGroup)
}

type Messager interface {
	ListenForResponseInfo(context.Context, *sync.WaitGroup, <-chan messaging.ResponseInfo)
	ListenForNodeEvents(context.Context, *sync.WaitGroup, <-chan registry.NodeEvent, chan error, string)
	Publish(context.Context, messaging.Message) error
	Response(context.Context, messaging.Message) error
	Request(context.Context, messaging.Message) error
}

type MessageReceiver interface {
	SubscribeToSubject(string) <-chan messaging.Message
	ResponseChannel() <-chan messaging.ResponseInfo
}

type ErrorHandler interface {
	ReceiveErrors(chan error)
}

type CourierService struct {
	Id                  string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
	MessagerOptions     []client.ClientNodeOption
	NodeEventerOptions  []eventer.OperatorClientOption
	ErrorHandler        ErrorHandler
	NodeRegister
	Messager
	MessageReceiver
	NodeEventer
}

type CourierServiceOption func(courier *CourierService)

func WithId(id string) CourierServiceOption {
	return func(courier *CourierService) {
		courier.Id = id
	}
}

// Subscribes sets the subjects the CourierService services will listen for
func Subscribes(subjects ...string) CourierServiceOption {
	return func(courier *CourierService) {
		courier.SubscribedSubjects = subjects
	}
}

// Broadcasts sets the subjects CourierService services will produce on
func Broadcasts(subjects ...string) CourierServiceOption {
	return func(courier *CourierService) {
		courier.BroadcastedSubjects = subjects
	}
}

func WithNodeEventerOptions(options ...eventer.OperatorClientOption) CourierServiceOption {
	return func(courier *CourierService) {
		courier.NodeEventerOptions = append(courier.NodeEventerOptions, options...)
	}
}

func WithClientNodeOptions(options ...client.ClientNodeOption) CourierServiceOption {
	return func(courier *CourierService) {
		courier.MessagerOptions = append(courier.MessagerOptions, options...)
	}
}

func NewCourierService(ctx context.Context, options ...CourierServiceOption) (*CourierService, error) {
	courier := &CourierService{}

	for _, option := range options {
		option(courier)
	}

	disco, err := eventer.NewOperatorClient(courier.NodeEventerOptions...)
	if err != nil {
		return nil, fmt.Errorf("NewCourierService: %w", err)
	}

	courier.NodeRegister = registry.NewNodeRegistry()
	courier.MessageReceiver = server.NewMessagingServer()
	courier.NodeEventer = disco

	in := courier.EventInChannel()
	courier.Messager = client.NewMessagingClient(in, courier.MessagerOptions...)

	wg := &sync.WaitGroup{}
	wg.Add(4)
	errs := make(chan error)
	nodeEventListener := courier.SubscribeToEvents()
	responseChannel := courier.ResponseChannel()

	go courier.DiscoverNodeEvents(ctx, in, errs, wg)
	go courier.RegisterNodes(ctx, wg, errs)
	go courier.ListenForNodeEvents(ctx, wg, nodeEventListener, errs, courier.Id)
	go courier.ListenForResponseInfo(ctx, wg, responseChannel)
	go courier.ErrorHandler.ReceiveErrors(errs)

	return courier, nil
}

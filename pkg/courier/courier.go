package courier

import (
	"context"
	"sync"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/platform-edn/courier/pkg/server"
)

type NodeRegister interface {
	EventInChannel() chan registry.NodeEvent
	SubscribeToEvents() chan registry.NodeEvent
	RegisterNodes(context.Context, *sync.WaitGroup, chan error)
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

type CourierService struct {
	Id                  string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
	MessagerOptions     []client.ClientNodeOption
	NodeRegister
	Messager
	MessageReceiver
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

	courier.NodeRegister = registry.NewNodeRegistry()
	courier.MessageReceiver = server.NewMessagingServer()

	in := courier.EventInChannel()
	courier.Messager = client.NewMessagingClient(in, courier.MessagerOptions...)

	wg := &sync.WaitGroup{}
	wg.Add(3)
	errchan := make(chan error)
	nodeEventListener := courier.SubscribeToEvents()
	responseChannel := courier.ResponseChannel()

	go courier.RegisterNodes(ctx, wg, errchan)
	go courier.ListenForNodeEvents(ctx, wg, nodeEventListener, errchan, courier.Id)
	go courier.ListenForResponseInfo(ctx, wg, responseChannel)

	return courier, nil
}

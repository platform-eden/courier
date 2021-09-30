package courier

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/observer"
	"github.com/platform-edn/courier/proto"
	"github.com/platform-edn/courier/proxy"
	"github.com/platform-edn/courier/server"
	"google.golang.org/grpc"
)

// CourierOptions to pass to NewCourier for creating a Courier struct
type CourierOptions struct {
	NodeStore           observer.NodeStorer
	ObserveInterval     time.Duration
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Address             string
	Port                string
}

// Courier is a messaging and node discovery service
type Courier struct {
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Address             string
	Port                string
	messageServer       *server.MessageServer
	messageProxy        *proxy.MessageProxy
	storeObserver       *observer.StoreObserver
}

// NewCourier creates a new Courier service
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

	push := make(chan message.Message)
	info := make(chan node.ResponseInfo)
	pquit := make(chan bool)

	mserver := server.NewMessageServer(push, info)
	err := startMessageServer(mserver, options.Port)
	if err != nil {
		return nil, fmt.Errorf("could not start message server: %s", err)
	}
	prox := proxy.NewMessageProxy(push, pquit)

	c := Courier{
		BroadcastedSubjects: options.BroadcastedSubjects,
		SubscribedSubjects:  options.SubscribedSubjects,
		Address:             options.Address,
		Port:                options.Port,
		messageServer:       mserver,
		messageProxy:        prox,
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

// startMessageServer starts the message server on a given port
func startMessageServer(m *server.MessageServer, port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("could not listen on port %s: %s", port, err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterMessageServerServer(grpcServer, m)

	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed serving: %s", err)
	}

	return nil
}

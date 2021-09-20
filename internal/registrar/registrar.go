package registrar

import (
	"errors"
	"fmt"
	"net"
)

type NodeRegistrar struct {
	Store               NodeStorer
	Registry            *NodeRegistry
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Port                string
}

func NewNodeRegistrar(store NodeStorer, options RegistrarOptions) (*NodeRegistrar, error) {
	if store == nil {
		return nil, errors.New("cannot create registrar without a nodeStorer")
	} else if options.Port == "" {
		return nil, errors.New("port cannot be blank")
	}

	// potential spot for concurrency, but will be synchronous for now
	ip, err := getLocalIp()
	if err != nil {
		return nil, err
	}

	node := NewNode(ip, options.Port)

	err = store.AddNode(node)
	if err != nil {
		return nil, fmt.Errorf("error adding node to store: %s", err)
	}

	subscribers, err := store.GetNodes(options.BroadcastedSubjects)
	if err != nil {
		return nil, fmt.Errorf("could not get subscribers: %s", err)
	}

	registry := NewNodeRegistry(node, subscribers)

	registrar := NodeRegistrar{
		Store:               store,
		Registry:            registry,
		BroadcastedSubjects: options.BroadcastedSubjects,
		SubscribedSubjects:  options.SubscribedSubjects,
		Port:                options.Port,
	}

	return &registrar, nil

}

// returns the current nodes ip that they will be receiving messages on
func getLocalIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("could not get local ip: %s", err)
	}
	defer conn.Close()

	address := conn.LocalAddr().(*net.UDPAddr)

	return address.IP.String(), nil
}

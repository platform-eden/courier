package registrar

import "github.com/google/uuid"

type Node struct {
	Id        string
	IpAddress string
	Port      string
}

func NewNode(ip string, port string) *Node {
	n := Node{
		Id:        uuid.NewString(),
		IpAddress: ip,
		Port:      port,
	}

	return &n
}

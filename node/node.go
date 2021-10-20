package node

type Node struct {
	Id                  string
	Address             string
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}

func NewNode(id string, address string, port string, subscribed []string, broadcasted []string) *Node {
	n := Node{
		Id:                  id,
		Address:             address,
		Port:                port,
		SubscribedSubjects:  subscribed,
		BroadcastedSubjects: broadcasted,
	}

	return &n
}

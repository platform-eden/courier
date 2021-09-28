package courier

type Node struct {
	Id                  string
	IpAddress           string
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}

func NewNode(id string, ip string, port string, subscribed []string, broadcasted []string) *Node {
	n := Node{
		Id:                  id,
		IpAddress:           ip,
		Port:                port,
		SubscribedSubjects:  subscribed,
		BroadcastedSubjects: broadcasted,
	}

	return &n
}

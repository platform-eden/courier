package courier

type node struct {
	Id                  string
	IpAddress           string
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}

func NewNode(id string, ip string, port string, subscribed []string, broadcasted []string) *node {
	n := node{
		Id:                  id,
		IpAddress:           ip,
		Port:                port,
		SubscribedSubjects:  subscribed,
		BroadcastedSubjects: broadcasted,
	}

	return &n
}

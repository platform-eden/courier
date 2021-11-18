package courier

type Noder interface {
	Id() string
	Address() string
	Subscribed() []string
	Broadcasted() []string
	Port() string
}

type Node struct {
	id                  string
	address             string
	port                string
	subscribedSubjects  []string
	broadcastedSubjects []string
}

func NewNode(id string, address string, port string, subscribed []string, broadcasted []string) *Node {
	n := Node{
		id:                  id,
		address:             address,
		port:                port,
		subscribedSubjects:  subscribed,
		broadcastedSubjects: broadcasted,
	}

	return &n
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) Address() string {
	return n.address
}

func (n *Node) Port() string {
	return n.port
}

func (n *Node) Subscribed() []string {
	return n.subscribedSubjects
}

func (n *Node) Broadcasted() []string {
	return n.broadcastedSubjects
}

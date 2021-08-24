package message

// metadata on a node that participated in the messaging
type Participant struct {
	Id      string
	Address string
	Port    string
	Sender  bool
}

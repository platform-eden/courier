package message

// the types of message that can be created
// assigned based on what kind interaction the sender of the message wants with receiver
type MessageType int

const (
	PubMessage MessageType = iota
	ReqMessage
	RespMessage
)

func (m MessageType) String() string {
	types := []string{
		"PubMessage",
		"RespMessage",
		"ReqMessage",
	}

	return types[int(m)]
}

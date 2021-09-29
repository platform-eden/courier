package message

// the types of message that can be created
// assigned based on what kind interaction the sender of the message wants with receiver
type messageType int

const (
	PubMessage messageType = iota
	ReqMessage
	RespMessage
)

func (m messageType) String() string {
	types := []string{
		"PubMessage",
		"RespMessage",
		"ReqMessage",
	}

	return types[int(m)]
}

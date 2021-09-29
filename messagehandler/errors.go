package messagehandler

type emptySubscriptionError struct{}

func (m *emptySubscriptionError) Error() string {
	return "cannot sends message for subject with no subscribers"
}

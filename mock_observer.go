package courier

type MockObserver struct {
	observeChannel chan []Noder
	fail           bool
}

func newMockObserver(ochan chan []Noder, fail bool) *MockObserver {
	o := MockObserver{
		observeChannel: ochan,
		fail:           fail,
	}

	return &o
}

func (o *MockObserver) Observe() (chan []Noder, error) {
	if o.fail {
		return nil, &ExpectedFailureError{}
	}

	return o.observeChannel, nil
}

func (o *MockObserver) AddNode(*Node) error {
	if o.fail {
		return &ExpectedFailureError{}
	}

	return nil
}

func (o *MockObserver) RemoveNode(*Node) error {
	if o.fail {
		return &ExpectedFailureError{}
	}

	return nil
}

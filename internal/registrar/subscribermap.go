package registrar

import "fmt"

type SubscriberMap struct {
	subscribers map[string][]*Node
	lock        *TicketLock
}

func NewSubscriberMap() *SubscriberMap {
	s := SubscriberMap{
		subscribers: make(map[string][]*Node),
		lock:        newTicketLock(),
	}

	return &s
}

// Goes through a node's subscribed subjects and updates the subscriber map if
// they aren't present under one of their subscribed subjects.
func (s *SubscriberMap) AddSubscriber(subscriber *Node) {
	s.lock.lock()
	defer s.lock.unlock()

	go func() {
		for _, subject := range subscriber.SubscribedSubjects {
			_, exists := findSubscriber(s.subscribers[subject], subscriber.Id)

			if !exists {
				s.subscribers[subject] = append(s.subscribers[subject], subscriber)
			}
		}
	}()
}

// Removes a subscriber from all of the subjects it subscribed to.
func (s *SubscriberMap) RemoveSubscriber(subscriber *Node) {
	s.lock.lock()
	defer s.lock.unlock()

	go func() {
		for _, subject := range subscriber.SubscribedSubjects {
			i, exists := findSubscriber(s.subscribers[subject], subscriber.Id)

			if exists {
				nodes, err := removeSubscriber(s.subscribers[subject], i)
				if err != nil {
					panic(err)
				}

				s.subscribers[subject] = nodes
			}
		}
	}()
}

// Takes a subject and returns all of nodes subscribed to that subject.
func (s *SubscriberMap) SubjectSubscribers(subject string) []*Node {
	s.lock.lock()
	defer s.lock.unlock()

	return s.subscribers[subject]
}

// Returns a unique list containing all of this nodes Subscribers.
func (s *SubscriberMap) GetAllSubscribers() []*Node {
	s.lock.lock()
	defer s.lock.unlock()

	var allsubs []*Node
	for _, subscribers := range s.subscribers {
		allsubs = append(allsubs, subscribers...)
	}

	return uniqueSubscribers(allsubs)
}

// checks if subscriber is listed under a subject.  If it exists, this returns the postion.
// If it doesn't, false is returned.
func findSubscriber(subscribers []*Node, id string) (int, bool) {
	for i, subscriber := range subscribers {
		if subscriber.Id == id {
			return i, true
		}
	}

	return -1, false
}

// Removes the node at the given position from the given node list.
func removeSubscriber(nodes []*Node, position int) ([]*Node, error) {
	if position < 0 || position > len(nodes) {
		return nil, fmt.Errorf("expected position to fall within range of %v but got %v", len(nodes), position)
	}
	nodes[position] = nodes[len(nodes)-1]
	return nodes[:len(nodes)-1], nil
}

// Takes a list of Nodes and returns a unique list
func uniqueSubscribers(subscribers []*Node) []*Node {
	subMap := make(map[string]*Node)
	for _, subscriber := range subscribers {
		subMap[subscriber.Id] = subscriber
	}

	var uniquesubs []*Node

	for _, subscriber := range subMap {
		uniquesubs = append(uniquesubs, subscriber)
	}

	return uniquesubs
}

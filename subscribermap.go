package courier

import (
	"fmt"
)

type subscriberMap struct {
	subscribers map[string][]*node
	lock        *ticketLock
}

func NewSubscriberMap() *subscriberMap {
	s := subscriberMap{
		subscribers: make(map[string][]*node),
		lock:        newTicketLock(),
	}

	return &s
}

// Goes through a node's subscribed subjects and updates the subscriber map if
// they aren't present under one of their subscribed subjects.
func (s *subscriberMap) AddSubscriber(subscriber *node) {
	s.lock.lock()
	defer s.lock.unlock()

	for _, subject := range subscriber.SubscribedSubjects {
		_, exists := findSubscriber(s.subscribers[subject], subscriber.Id)

		if !exists {
			s.subscribers[subject] = append(s.subscribers[subject], subscriber)
		}
	}
}

// Removes a subscriber from all of the subjects it subscribed to.
func (s *subscriberMap) RemoveSubscriber(subscriber *node) error {
	s.lock.lock()
	defer s.lock.unlock()

	for _, subject := range subscriber.SubscribedSubjects {
		i, exists := findSubscriber(s.subscribers[subject], subscriber.Id)

		if exists {
			if len(s.subscribers[subject]) == 1 {
				delete(s.subscribers, subject)
			} else {
				nodes, err := removeSubscriberFromSubject(s.subscribers[subject], i)
				if err != nil {
					return fmt.Errorf("could not remove subscriber from %s: %s", subject, err)
				}

				s.subscribers[subject] = nodes
			}
		}
	}

	return nil
}

// Takes a subject and returns all of nodes subscribed to that subject.
func (s *subscriberMap) SubjectSubscribers(subject string) ([]*node, error) {
	s.lock.lock()
	defer s.lock.unlock()

	subscribers, ok := s.subscribers[subject]
	if !ok {
		return nil, fmt.Errorf("subject %s does not exist", subject)
	}

	return subscribers, nil
}

// Returns a unique list containing all of this nodes Subscribers.
func (s *subscriberMap) AllSubscribers() []*node {
	s.lock.lock()
	defer s.lock.unlock()

	uniquesubs := []*node{}
	subMap := make(map[string]bool)
	for _, subscribers := range s.subscribers {
		for _, subscriber := range subscribers {
			_, ok := subMap[subscriber.Id]
			if !ok {
				subMap[subscriber.Id] = true
				uniquesubs = append(uniquesubs, subscriber)
			}
		}
	}

	return uniquesubs
}

// checks if subscriber is listed under a subject.  If it exists, this returns the postion.
// If it doesn't, false is returned.
func findSubscriber(subscribers []*node, id string) (int, bool) {
	for i, subscriber := range subscribers {
		if subscriber.Id == id {
			return i, true
		}
	}

	return -1, false
}

// Removes the node at the given position from the given node list.
func removeSubscriberFromSubject(nodes []*node, position int) ([]*node, error) {
	if position < 0 || position > len(nodes) {
		return nil, fmt.Errorf("expected position to fall within range of %v but got %v", len(nodes), position)
	}
	nodes[position] = nodes[len(nodes)-1]
	return nodes[:len(nodes)-1], nil
}

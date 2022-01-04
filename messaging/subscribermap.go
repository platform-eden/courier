package messaging

import (
	"sort"
)

type subscriberMap struct {
	subjectSubscribers map[string][]string
	lock               *TicketLock
}

func newSubscriberMap() *subscriberMap {
	s := subscriberMap{
		subjectSubscribers: map[string][]string{},
		lock:               NewTicketLock(),
	}

	return &s
}

func (s *subscriberMap) Add(id string, subjects ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, subject := range subjects {
		s.subjectSubscribers[subject] = append(s.subjectSubscribers[subject], id)
	}
}

func (s *subscriberMap) Remove(id string, subjects ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, subject := range subjects {
		ids := s.subjectSubscribers[subject]

		sort.Strings(ids)
		position := sort.SearchStrings(ids, id)
		if position == len(ids) {
			continue
		}

		ids[position] = ids[len(ids)-1]
		ids[len(ids)-1] = ""
		ids = ids[:len(ids)-1]

		s.subjectSubscribers[subject] = ids
	}
}

// Subscribers takes a subject and returns all of nodes subscribed to that subject.
func (s *subscriberMap) Subscribers(subject string) ([]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ids, ok := s.subjectSubscribers[subject]
	if !ok {
		return nil, &UnregisteredSubscriberError{
			Method:  "Subscribers",
			Subject: subject,
		}
	}

	return ids, nil
}

// CheckForSubscriber sees if a node id exists in a subject.  Returns true if it exists.
// Only used for testing.
func (s *subscriberMap) CheckForSubscriber(subject string, id string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	subIds, exist := s.subjectSubscribers[subject]
	if !exist {
		return false
	}

	for _, subId := range subIds {
		if id == subId {
			return true
		}
	}

	return false
}

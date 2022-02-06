package client

import (
	"fmt"
	"sort"

	"github.com/platform-edn/courier/pkg/lock"
)

type subscriberMap struct {
	SubjectSubscribers map[string][]string
	lock.Locker
}

func NewSubscriberMap() *subscriberMap {
	s := subscriberMap{
		SubjectSubscribers: map[string][]string{},
		Locker:             lock.NewTicketLock(),
	}

	return &s
}

func (s *subscriberMap) AddSubscriber(id string, subjects ...string) {
	s.Lock()
	defer s.Unlock()

	for _, subject := range subjects {
		s.SubjectSubscribers[subject] = append(s.SubjectSubscribers[subject], id)
	}
}

func (s *subscriberMap) RemoveSubscriber(id string, subjects ...string) {
	s.Lock()
	defer s.Unlock()

	for _, subject := range subjects {
		ids := s.SubjectSubscribers[subject]

		sort.Strings(ids)
		position := sort.SearchStrings(ids, id)
		if position == len(ids) {
			continue
		}

		ids[position] = ids[len(ids)-1]
		ids[len(ids)-1] = ""
		ids = ids[:len(ids)-1]

		s.SubjectSubscribers[subject] = ids
	}
}

// Subscribers takes a subject and returns all of nodes subscribed to that subject.
func (s *subscriberMap) Subscribers(subject string) ([]string, error) {
	s.Lock()
	defer s.Unlock()

	ids, ok := s.SubjectSubscribers[subject]
	if !ok {
		return nil, fmt.Errorf("Subscribers: %w", &UnregisteredSubjectError{
			Subject: subject,
		})
	}

	return ids, nil
}

func (s *subscriberMap) GenerateIdsBySubject(subject string) (<-chan string, error) {
	out := make(chan string)

	ids, err := s.Subscribers(subject)
	if err != nil {
		return nil, fmt.Errorf("GenerateIdsBySubject: %w", err)
	}

	go func() {
		for _, id := range ids {
			out <- id
		}
		close(out)
	}()

	return out, nil
}

// CheckForSubscriber sees if a node id exists in a subject.  Returns true if it exists.
// Only used for testing.
// func (s *subscriberMap) CheckForSubscriber(subject string, id string) bool {
// 	s.Lock()
// 	defer s.Unlock()

// 	subIds, exist := s.SubjectSubscribers[subject]
// 	if !exist {
// 		return false
// 	}

// 	for _, subId := range subIds {
// 		if id == subId {
// 			return true
// 		}
// 	}

// 	return false
// }

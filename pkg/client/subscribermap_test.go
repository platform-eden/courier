package client_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestUnregisteredSubscriberError_Error(t *testing.T) {
	method := "testMethod"
	subject := "test"
	e := &client.UnregisteredSubscriberError{
		Method:  method,
		Subject: subject,
	}

	message := e.Error()

	if message != fmt.Sprintf("%s: no subscribers registered for subject %s", method, subject) {
		t.Fatalf("expected error message to be %s but got %s", fmt.Sprintf("%s: no subscribers registered for subject %s", method, subject), message)
	}
}

func TestSubscriberMap_Add(t *testing.T) {
	tests := map[string]struct {
		subscribers []string
		subjects    []string
	}{
		"adds all ids to given subjects": {
			subscribers: []string{"1", "2", "3"},
			subjects:    []string{"sub1", "sub2", "sub3"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			subMap := client.NewSubscriberMap()
			assert := assert.New(t)

			for _, id := range test.subscribers {
				subMap.AddSubscriber(id, test.subjects...)
			}

			assert.Equal(len(subMap.SubjectSubscribers), len(test.subjects))

			for _, v := range subMap.SubjectSubscribers {
				assert.EqualValues(v, test.subscribers)
			}
		})
	}
}

func TestSubscriberMap_Remove(t *testing.T) {
	tests := map[string]struct {
		addSubscribers    []string
		removeSubscribers []string
		subjects          []string
	}{
		"all added subscribers are removed": {
			addSubscribers:    []string{"1", "2", "3"},
			removeSubscribers: []string{"1", "2", "3"},
			subjects:          []string{"sub1", "sub2", "sub3"},
		},
		"removing a subscriber that doesn't exist doesn't panic": {
			addSubscribers:    []string{"1", "2", "3"},
			removeSubscribers: []string{"1", "2", "3", "4"},
			subjects:          []string{"sub1", "sub2", "sub3"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			subMap := client.NewSubscriberMap()

			for _, id := range test.addSubscribers {
				subMap.AddSubscriber(id, test.subjects...)
			}

			for _, id := range test.removeSubscribers {
				subMap.RemoveSubscriber(id, test.subjects...)
			}

			assert.Equal(len(test.subjects), len(subMap.SubjectSubscribers))
			for _, v := range subMap.SubjectSubscribers {
				assert.Zero(len(v))
			}
		})
	}
}

func TestSubscriberMap_Subscribers(t *testing.T) {
	tests := map[string]struct {
		creationSubject string
		accessSubject   string
		subscribers     []string
		err             error
	}{
		"returns all subscribers for a subject": {
			creationSubject: "subject",
			accessSubject:   "subject",
			subscribers:     []string{"1", "2", "3"},
			err:             nil,
		},
		"errors when accessing an unregistered subject": {
			creationSubject: "subject",
			accessSubject:   "fail",
			subscribers:     []string{"1", "2", "3"},
			err:             &client.UnregisteredSubjectError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			subMap := client.NewSubscriberMap()
			subMap.SubjectSubscribers[test.creationSubject] = test.subscribers

			subscribers, err := subMap.Subscribers(test.accessSubject)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.EqualValues(test.subscribers, subscribers)
		})
	}
}

func TestSubscriberMap_GenerateIdsBySubject(t *testing.T) {
	tests := map[string]struct {
		creationSubject string
		accessSubject   string
		subscribers     []string
		err             error
	}{
		"returns a channel with all subscribers for a subject": {
			creationSubject: "subject",
			accessSubject:   "subject",
			subscribers:     []string{"1", "2", "3"},
			err:             nil,
		},
		"errors when accessing an unregistered subject": {
			creationSubject: "subject",
			accessSubject:   "fail",
			subscribers:     []string{"1", "2", "3"},
			err:             &client.UnregisteredSubjectError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			subMap := client.NewSubscriberMap()

			subMap.SubjectSubscribers[test.creationSubject] = test.subscribers

			out, err := subMap.GenerateIdsBySubject(test.accessSubject)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)

			done := make(chan struct{})
			go func() {
				subOut := []string{}
				for s := range out {
					subOut = append(subOut, s)
				}

				assert.EqualValues(test.subscribers, subOut)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second * 3):
				t.Fatal("did not close done channel in time")
			}
		})
	}
}

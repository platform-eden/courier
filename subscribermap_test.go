package courier

import (
	"testing"

	"github.com/google/uuid"
)

/**************************************************************
Expected Outcomes:
- the id passed in should be added to a list for each subject passed in
**************************************************************/
func TestSubscriberMap_Add(t *testing.T) {
	type test struct {
		nodeCount int
	}

	tests := []test{
		{
			nodeCount: 10,
		},
	}

	for _, tc := range tests {
		subMap := newSubscriberMap()

		for i := 0; i < tc.nodeCount; i++ {
			id := uuid.NewString()
			subjects := []string{"sub1", "sub2", "sub3"}

			subMap.Add(id, subjects...)
		}

		if len(subMap.subjectSubscribers) != 3 {
			t.Fatalf("expected 3 subjects but got %v", len(subMap.subjectSubscribers))
		}

		for _, v := range subMap.subjectSubscribers {
			if len(v) != tc.nodeCount {
				t.Fatalf("expected subject to have %v ids but got %v instead", tc.nodeCount, len(v))
			}
		}
	}
}

/**************************************************************
Expected Outcomes:
- all ids passed in should be removed from the given subjects
**************************************************************/
func TestSubscriberMap_Remove(t *testing.T) {
	type test struct {
		nodeCount int
		exists    bool
	}

	tests := []test{
		{
			nodeCount: 10,
			exists:    true,
		},
		{
			nodeCount: 10,
			exists:    false,
		},
	}

	for _, tc := range tests {
		subMap := newSubscriberMap()
		ids := []string{}
		subjects := []string{"sub1", "sub2", "sub3"}

		for i := 0; i < tc.nodeCount; i++ {
			id := uuid.NewString()
			ids = append(ids, id)

			subMap.Add(id, subjects...)
		}

		if !tc.exists {
			subMap.Remove("badId", subjects...)
		}

		for _, id := range ids {
			subMap.Remove(id, subjects...)
		}

		if len(subMap.subjectSubscribers) != 3 {
			t.Fatalf("expected 3 subjects but got %v", len(subMap.subjectSubscribers))
		}

		for _, v := range subMap.subjectSubscribers {
			if len(v) != 0 {
				t.Fatalf("expected subject to have %v ids but got %v instead", tc.nodeCount, len(v))
			}
		}
	}
}

/**************************************************************
Expected Outcomes:
- ids passed in should be removed from all subjects passed in
**************************************************************/
func TestSubscriberMap_Subscribers(t *testing.T) {
	type test struct {
		subject         string
		expectedFailure bool
		nodeCount       int
	}

	tests := []test{
		{
			subject:         "test",
			expectedFailure: true,
			nodeCount:       10,
		},
		{
			subject:         "test",
			expectedFailure: false,
			nodeCount:       10,
		},
	}

	for _, tc := range tests {
		subMap := newSubscriberMap()

		for i := 0; i < tc.nodeCount; i++ {
			id := uuid.NewString()

			subMap.Add(id, tc.subject)
		}

		if tc.expectedFailure {
			_, err := subMap.Subscribers("failure")
			if err != nil {
				continue
			}

			t.Fatal("expected Subscribers to fail but it didn't")
		}

		subscribers, err := subMap.Subscribers(tc.subject)
		if err != nil {
			t.Fatalf("expected Subscribers to pass but it failed: %s", err)
		}

		if len(subscribers) != tc.nodeCount {
			t.Fatalf("expected subject length to be %v but got %v", tc.nodeCount, len(subscribers))
		}
	}
}

/**************************************************************
Expected Outcomes:
- if an id exists in a subject list, this should return true
- if a subject doesn't exist, this should return false
**************************************************************/
func TestSubscriberMap_CheckForSubscriber(t *testing.T) {
	type test struct {
		subject         string
		checkSubject    string
		id              string
		checkId         string
		expectedOutcome bool
	}

	tests := []test{
		{
			subject:         "test",
			checkSubject:    "wowowow",
			id:              "test",
			checkId:         "test",
			expectedOutcome: false,
		},
		{
			subject:         "test",
			checkSubject:    "test",
			id:              "test",
			checkId:         "test",
			expectedOutcome: true,
		},
		{
			subject:         "test",
			checkSubject:    "test",
			id:              "test",
			checkId:         "test1",
			expectedOutcome: false,
		},
	}

	for _, tc := range tests {
		subMap := newSubscriberMap()

		subMap.Add(tc.subject, tc.id)

		exist := subMap.CheckForSubscriber(tc.checkSubject, tc.checkId)

		if exist && !tc.expectedOutcome {
			t.Fatalf("expected to return false but it returned true")
		}

		if !exist && tc.expectedOutcome {
			t.Fatalf("expected to return true but it returned false")
		}
	}
}

package server

import (
	"testing"
)

/**************************************************************
Expected Outcomes:
- subject in chanMap should have a channel for every time Add is called
// **************************************************************/
func TestChannelMap_Add(t *testing.T) {
	type test struct {
		channelCount int
		subject      string
	}

	tests := []test{
		{
			channelCount: 10,
			subject:      "test",
		},
	}

	for _, tc := range tests {
		chanMap := newChannelMap()
		defer chanMap.CloseSubscriberChannels()

		for i := 0; i < tc.channelCount; i++ {
			chanMap.SubscribeToSubject(tc.subject)
		}

		if len(chanMap.SubjectChannels[tc.subject]) != tc.channelCount {
			t.Fatalf("expected length of %s channels to be %v but got %v", tc.subject, tc.channelCount, len(chanMap.SubjectChannels[tc.subject]))
		}
	}
}

/**************************************************************
Expected Outcomes:
- should return a list of channels with a length equal to how many adds were done on the subject passed in
- should return error if subject passed in isn't registered
**************************************************************/
func TestChannelMap_Subscriptions(t *testing.T) {
	type test struct {
		addSubject          string
		subscriptionSubject string
		channelCount        int
		expectedFailure     bool
	}

	tests := []test{
		{
			addSubject:          "test",
			subscriptionSubject: "test",
			channelCount:        10,
			expectedFailure:     false,
		},
		{
			addSubject:          "test",
			subscriptionSubject: "test1",
			channelCount:        10,
			expectedFailure:     true,
		},
	}

	for _, tc := range tests {
		chanMap := newChannelMap()
		defer chanMap.CloseSubscriberChannels()

		for i := 0; i < tc.channelCount; i++ {
			chanMap.SubscribeToSubject(tc.addSubject)
		}

		channels, err := chanMap.Subscriptions(tc.subscriptionSubject)
		if err != nil && !tc.expectedFailure {
			t.Fatalf("Subscriptions failed when it shouldn't have: %s", err)
		}

		if len(channels) != tc.channelCount && !tc.expectedFailure {
			t.Fatalf("expected length of %s channels to be %v but got %v", tc.addSubject, tc.channelCount, len(channels))
		}

	}
}

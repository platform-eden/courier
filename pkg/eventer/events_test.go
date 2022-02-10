package eventer_test

import (
	"testing"

	"github.com/platform-edn/courier/pkg/eventer"
	"github.com/stretchr/testify/assert"
)

func TestNewEventTypeMap(t *testing.T) {
	assert := assert.New(t)
	eventTypeMap := eventer.NewEventTypeMap()

	assert.Len(eventTypeMap, 3)
}

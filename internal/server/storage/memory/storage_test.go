package memory

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeSetGet(t *testing.T) {
	s := NewStorage()

	metricName := "foo"
	s.GaugeSet(metricName, 19.17)
	val, exists := s.GaugeGet(metricName)
	assert.True(t, exists)
	assert.Equal(t, 19.17, val)
}

func TestCounterSetGet(t *testing.T) {
	s := NewStorage()

	metricName := "bar"
	s.CounterSet(metricName, 1)
	val, exists := s.CounterGet(metricName)
	assert.True(t, exists)
	assert.Equal(t, int64(1), val)
}

func TestCounterGetNotExists(t *testing.T) {
	s := NewStorage()

	metricName := "not_found"
	val, exists := s.CounterGet(metricName)
	assert.False(t, exists)
	assert.Equal(t, int64(0), val)
}

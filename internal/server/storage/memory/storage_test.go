package memory

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeSetGet(t *testing.T) {
	s := NewStorage()

	metricName := "foo"
	err := s.GaugeSet(metricName, 19.17)
	assert.NoError(t, err)

	val, err := s.GaugeGet(metricName)
	assert.NoError(t, err)
	assert.Equal(t, 19.17, val)
}

func TestCounterSetGet(t *testing.T) {
	s := NewStorage()

	metricName := "bar"
	err := s.CounterInc(metricName, 1)
	assert.NoError(t, err)

	val, err := s.CounterGet(metricName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestCounterGetNotExists(t *testing.T) {
	s := NewStorage()

	metricName := "not_found"
	val, err := s.CounterGet(metricName)
	assert.Error(t, err)
	assert.Equal(t, int64(0), val)
}

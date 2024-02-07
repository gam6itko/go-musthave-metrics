package memory

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeSetGet(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "foo"
	err := s.GaugeSet(ctx, metricName, 19.17)
	assert.NoError(t, err)

	val, err := s.GaugeGet(ctx, metricName)
	assert.NoError(t, err)
	assert.Equal(t, 19.17, val)
}

func TestCounterSetGet(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "bar"
	err := s.CounterInc(ctx, metricName, 1)
	assert.NoError(t, err)

	val, err := s.CounterGet(ctx, metricName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestCounterGetNotExists(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "not_found"
	val, err := s.CounterGet(ctx, metricName)
	assert.Error(t, err)
	assert.Equal(t, int64(0), val)
}

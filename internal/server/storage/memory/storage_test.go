package memory

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGaugeSetGet(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "foo"
	err := s.GaugeSet(ctx, metricName, 19.17)
	require.NoError(t, err)

	val, err := s.GaugeGet(ctx, metricName)
	require.NoError(t, err)
	require.InDelta(t, 19.17, val, .0001)
}

func TestCounterSetGet(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "bar"
	err := s.CounterInc(ctx, metricName, 1)
	require.NoError(t, err)

	val, err := s.CounterGet(ctx, metricName)
	require.NoError(t, err)
	require.Equal(t, int64(1), val)
}

func TestCounterGetNotExists(t *testing.T) {
	s := NewStorage()

	ctx := context.Background()

	metricName := "not_found"
	val, err := s.CounterGet(ctx, metricName)
	require.Error(t, err)
	require.InDelta(t, 0, val, .1)
}

package memory

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeSetGet(t *testing.T) {
	metricName := "foo"
	GaugeSet(metricName, 19.17)
	val, exists := GaugeGet(metricName)
	assert.True(t, exists)
	assert.Equal(t, 19.17, val)
}

func TestCounterSetGet(t *testing.T) {
	metricName := "bar"
	CounterInc(metricName, 1)
	val, exists := CounterGet(metricName)
	assert.True(t, exists)
	assert.Equal(t, int64(1), val)
}

func TestCounterGetNotExists(t *testing.T) {
	metricName := "not_found"
	val, exists := CounterGet(metricName)
	assert.False(t, exists)
	assert.Equal(t, int64(0), val)
}

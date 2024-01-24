package retrible

import (
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Decoration(t *testing.T) {
	inner := memory.NewStorage()
	s := NewStorage(inner)

	metricName := "foo"
	err := s.GaugeSet(metricName, 19.17)
	assert.NoError(t, err)

	val, err := s.GaugeGet(metricName)
	assert.NoError(t, err)

	assert.Equal(t, 19.17, val)
}

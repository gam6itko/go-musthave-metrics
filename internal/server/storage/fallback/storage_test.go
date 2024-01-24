package fallback

import (
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/server/mocks"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Decoration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inner := mocks.NewMockStorage(ctrl)
	// counter
	inner.EXPECT().
		CounterInc("foo", int64(1)).
		Return(errors.New("boom"))
	inner.EXPECT().
		CounterGet("foo").
		Return(int64(0), errors.New("boom"))
	inner.EXPECT().
		CounterAll().
		Return(map[string]int64{}, errors.New("boom"))
	// gauge
	inner.EXPECT().
		GaugeSet("bar", 19.17).
		Return(errors.New("boom"))
	inner.EXPECT().
		GaugeGet("bar").
		Return(float64(0.0), errors.New("boom"))
	inner.EXPECT().
		GaugeAll().
		Return(map[string]float64{}, errors.New("boom"))

	fallback := memory.NewStorage()

	s := NewStorage(inner, fallback)

	//counter
	err := s.CounterInc("foo", 1)
	assert.NoError(t, err)
	counterVal, err := s.CounterGet("foo")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), counterVal)
	counterAll, err := s.CounterAll()
	assert.NoError(t, err)
	assert.Len(t, counterAll, 1)
	// gauge
	err = s.GaugeSet("bar", 19.17)
	assert.NoError(t, err)
	gaugeVal, err := s.GaugeGet("bar")
	assert.NoError(t, err)
	assert.Equal(t, float64(19.17), gaugeVal)
	gaugeAll, err := s.GaugeAll()
	assert.NoError(t, err)
	assert.Len(t, gaugeAll, 1)
}

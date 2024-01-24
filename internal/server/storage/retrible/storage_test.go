package retrible

import (
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/server/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Decoration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//gomock.InOrder(
	//   m.EXPECT().Get("1"),
	//   m.EXPECT().Get("2"),
	//   m.EXPECT().Get("3"),
	//   m.EXPECT().Get("4"),
	//)

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

	s := NewStorage(
		inner,
		[]time.Duration{
			time.Nanosecond,
			time.Nanosecond,
		},
	)

	metricName := "foo"
	err := s.GaugeSet(metricName, 19.17)
	assert.NoError(t, err)

	val, err := s.GaugeGet(metricName)
	assert.NoError(t, err)

	assert.Equal(t, 19.17, val)
}

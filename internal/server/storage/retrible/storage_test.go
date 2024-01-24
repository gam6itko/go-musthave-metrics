package retrible

import (
	"github.com/gam6itko/go-musthave-metrics/internal/server/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Decoration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inner := mocks.NewMockStorage(ctrl)
	// counter
	gomock.InOrder(
		inner.EXPECT().
			CounterInc("counter1", int64(1)).
			Return(Error{}),
		inner.EXPECT().
			CounterInc("counter1", int64(1)).
			Return(Error{}),
		inner.EXPECT().
			CounterInc("counter1", int64(1)).
			Return(nil),
	)

	gomock.InOrder(
		inner.EXPECT().
			CounterGet("counter1").
			Return(int64(0), Error{}),
		inner.EXPECT().
			CounterGet("counter1").
			Return(int64(0), Error{}),
		inner.EXPECT().
			CounterGet("counter1").
			Return(int64(1), nil),
	)

	s := NewStorage(
		inner,
		[]time.Duration{
			time.Nanosecond,
			time.Nanosecond,
		},
	)

	err := s.CounterInc("counter1", 1)
	assert.NoError(t, err)

	val, err := s.CounterGet("counter1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

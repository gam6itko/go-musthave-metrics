package retrible

import (
	"context"
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/server/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// Вызвать дочерний storage 2 раза
func Test_Decoration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inner := mocks.NewMockStorage(ctrl)
	// counter
	inner.EXPECT().
		CounterInc(context.Background(), "foo", int64(1)).
		Return(errors.New("boom")).
		Times(3)
	inner.EXPECT().
		CounterGet(context.Background(), "foo").
		Return(int64(0), errors.New("boom")).
		Times(3)
	inner.EXPECT().
		CounterAll(context.Background()).
		Return(map[string]int64{}, errors.New("boom")).
		Times(3)
	// gauge
	inner.EXPECT().
		GaugeSet(context.Background(), "bar", 19.17).
		Return(errors.New("boom")).
		MinTimes(3)
	inner.EXPECT().
		GaugeGet(context.Background(), "bar").
		Return(float64(0.0), errors.New("boom")).
		MinTimes(3)
	inner.EXPECT().
		GaugeAll(context.Background()).
		Return(map[string]float64{}, errors.New("boom")).
		Times(3)

	s := NewStorage(
		inner,
		[]time.Duration{
			time.Nanosecond,
			time.Nanosecond,
		},
	)

	t.Run("counter", func(t *testing.T) {
		ctx := context.Background()

		name := "foo"
		err := s.CounterInc(ctx, name, 1)
		require.Error(t, err)

		_, err = s.CounterGet(ctx, name)
		require.Error(t, err)

		_, err = s.CounterAll(ctx)
		require.Error(t, err)
	})

	t.Run("gauge", func(t *testing.T) {
		ctx := context.Background()

		name := "bar"
		err := s.GaugeSet(ctx, name, 19.17)
		require.Error(t, err)

		_, err = s.GaugeGet(ctx, name)
		require.Error(t, err)

		_, err = s.GaugeAll(ctx)
		require.Error(t, err)
	})
}

func Test_Decoration2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inner := mocks.NewMockStorage(ctrl)
	// counter
	gomock.InOrder(
		inner.EXPECT().
			CounterInc(context.Background(), "counter1", int64(1)).
			Return(Error{}),
		inner.EXPECT().
			CounterInc(context.Background(), "counter1", int64(1)).
			Return(Error{}),
		inner.EXPECT().
			CounterInc(context.Background(), "counter1", int64(1)).
			Return(nil),
	)

	gomock.InOrder(
		inner.EXPECT().
			CounterGet(context.Background(), "counter1").
			Return(int64(0), Error{}),
		inner.EXPECT().
			CounterGet(context.Background(), "counter1").
			Return(int64(0), Error{}),
		inner.EXPECT().
			CounterGet(context.Background(), "counter1").
			Return(int64(1), nil),
	)

	s := NewStorage(
		inner,
		[]time.Duration{
			time.Nanosecond,
			time.Nanosecond,
		},
	)

	ctx := context.Background()

	err := s.CounterInc(ctx, "counter1", 1)
	require.NoError(t, err)

	val, err := s.CounterGet(ctx, "counter1")
	require.NoError(t, err)
	require.Equal(t, int64(1), val)
}

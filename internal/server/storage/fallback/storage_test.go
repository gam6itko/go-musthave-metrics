package fallback

import (
	"context"
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/server/mocks"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Decoration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	inner := mocks.NewMockStorage(ctrl)
	// counter
	inner.EXPECT().
		CounterInc(ctx, "foo", int64(1)).
		Return(errors.New("boom"))
	inner.EXPECT().
		CounterGet(ctx, "foo").
		Return(int64(0), errors.New("boom"))
	inner.EXPECT().
		CounterAll(ctx).
		Return(map[string]int64{}, errors.New("boom"))
	// gauge
	inner.EXPECT().
		GaugeSet(ctx, "bar", 19.17).
		Return(errors.New("boom"))
	inner.EXPECT().
		GaugeGet(ctx, "bar").
		Return(float64(0.0), errors.New("boom"))
	inner.EXPECT().
		GaugeAll(ctx).
		Return(map[string]float64{}, errors.New("boom"))

	fallback := memory.NewStorage()

	s := NewStorage(inner, fallback)

	//counter
	err := s.CounterInc(ctx, "foo", 1)
	require.NoError(t, err)
	counterVal, err := s.CounterGet(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), counterVal)
	counterAll, err := s.CounterAll(ctx)
	require.NoError(t, err)
	require.Len(t, counterAll, 1)
	// gauge
	err = s.GaugeSet(ctx, "bar", 19.17)
	require.NoError(t, err)
	gaugeVal, err := s.GaugeGet(ctx, "bar")
	require.NoError(t, err)
	require.InDelta(t, 19.17, gaugeVal, .0001)
	gaugeAll, err := s.GaugeAll(ctx)
	require.NoError(t, err)
	require.Len(t, gaugeAll, 1)
}

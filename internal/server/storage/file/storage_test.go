package file

import (
	"context"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"testing"
)

func Test_Storage_SaveLoad(t *testing.T) {

	filePath := fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("save not sync", func(t *testing.T) {
		ms := memory.NewStorage()

		ctx := context.Background()

		s, err := NewStorage(ms, filePath, false)
		require.NoError(t, err)
		err = s.CounterInc(ctx, "counter1", 1)
		require.NoError(t, err)
		err = s.GaugeSet(ctx, "gauge1", 2.2)
		require.NoError(t, err)
		err = s.Save(context.TODO())
		require.NoError(t, err)

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"counter":{"counter1":1},"gauge":{"gauge1":2.2}}`,
			string(b),
		)
		err = s.Close()
		require.NoError(t, err)
	})
	err := os.Remove(filePath)
	require.NoError(t, err)

	filePath = fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("save sync", func(t *testing.T) {
		ms := memory.NewStorage()

		ctx := context.Background()

		s, err := NewStorage(ms, filePath, true)
		require.NoError(t, err)
		err = s.CounterInc(ctx, "counter1", 3)
		require.NoError(t, err)
		err = s.GaugeSet(ctx, "gauge1", 4.4)
		require.NoError(t, err)

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"counter":{"counter1":3},"gauge":{"gauge1":4.4}}`,
			string(b),
		)

		err = s.Close()
		require.NoError(t, err)
	})
	err = os.Remove(filePath)
	require.NoError(t, err)

	filePath = fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("multi load", func(t *testing.T) {
		ctx := context.Background()

		ms := memory.NewStorage()
		err := ms.CounterInc(ctx, "counter3", 3)
		require.NoError(t, err)
		err = ms.GaugeSet(ctx, "gauge4", 4.4)
		require.NoError(t, err)

		s, err := NewStorage(ms, filePath, false)
		require.NoError(t, err)
		err = s.Save(context.TODO())
		require.NoError(t, err)

		ms2 := memory.NewStorage()
		s2, err := NewStorage(ms2, filePath, false)
		require.NoError(t, err)
		err = s2.Load()
		require.NoError(t, err)

		cVal1, err := ms.CounterGet(ctx, "counter3")
		require.NoError(t, err)
		cVal2, err := ms2.CounterGet(ctx, "counter3")
		require.NoError(t, err)
		require.Equal(t, cVal1, cVal2)

		gVal1, err := ms.GaugeGet(ctx, "gauge4")
		require.NoError(t, err)
		gVal2, err := ms2.GaugeGet(ctx, "gauge4")
		require.NoError(t, err)
		require.Equal(t, gVal1, gVal2)

		// еще один load
		err = ms2.CounterInc(ctx, "counter3", 9999)
		require.NoError(t, err)
		err = ms2.GaugeSet(ctx, "gauge4", 9999)
		require.NoError(t, err)
		err = s2.Load()
		require.NoError(t, err)

		cVal2, err = ms2.CounterGet(ctx, "counter3")
		require.NoError(t, err)
		require.Equal(t, int64(3), cVal2)
		gVal2, err = ms2.GaugeGet(ctx, "gauge4")
		require.NoError(t, err)
		require.Equal(t, 4.4, gVal2)
	})

	err = os.Remove(filePath)
	require.NoError(t, err)
}

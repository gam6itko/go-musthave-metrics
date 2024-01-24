package file

import (
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

		s, err := NewStorage(ms, filePath, false)
		require.NoError(t, err)
		err = s.CounterInc("counter1", 1)
		require.NoError(t, err)
		err = s.GaugeSet("gauge1", 2.2)
		require.NoError(t, err)
		err = s.Save()
		require.NoError(t, err)

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"Counter":{"counter1":1},"Gauge":{"gauge1":2.2}}`,
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

		s, err := NewStorage(ms, filePath, true)
		require.NoError(t, err)
		err = s.CounterInc("counter1", 3)
		require.NoError(t, err)
		err = s.GaugeSet("gauge1", 4.4)
		require.NoError(t, err)

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"Counter":{"counter1":3},"Gauge":{"gauge1":4.4}}`,
			string(b),
		)

		err = s.Close()
		require.NoError(t, err)
	})
	err = os.Remove(filePath)
	require.NoError(t, err)

	filePath = fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("multi load", func(t *testing.T) {
		ms := memory.NewStorage()
		err := ms.CounterInc("counter3", 3)
		require.NoError(t, err)
		err = ms.GaugeSet("gauge4", 4.4)
		require.NoError(t, err)

		s, err := NewStorage(ms, filePath, false)
		require.NoError(t, err)
		err = s.Save()
		require.NoError(t, err)

		ms2 := memory.NewStorage()
		s2, err := NewStorage(ms2, filePath, false)
		require.NoError(t, err)
		err = s2.Load()
		require.NoError(t, err)

		cVal1, err := ms.CounterGet("counter3")
		require.NoError(t, err)
		cVal2, err := ms2.CounterGet("counter3")
		require.NoError(t, err)
		require.Equal(t, cVal1, cVal2)

		gVal1, err := ms.GaugeGet("gauge4")
		require.NoError(t, err)
		gVal2, err := ms2.GaugeGet("gauge4")
		require.NoError(t, err)
		require.Equal(t, gVal1, gVal2)

		// еще один load
		err = ms2.CounterInc("counter3", 9999)
		require.NoError(t, err)
		err = ms2.GaugeSet("gauge4", 9999)
		require.NoError(t, err)
		err = s2.Load()
		require.NoError(t, err)

		cVal2, err = ms2.CounterGet("counter3")
		require.NoError(t, err)
		require.Equal(t, int64(3), cVal2)
		gVal2, err = ms2.GaugeGet("gauge4")
		require.NoError(t, err)
		require.Equal(t, 4.4, gVal2)
	})

	err = os.Remove(filePath)
	require.NoError(t, err)
}

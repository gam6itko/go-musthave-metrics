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

		s := NewStorage(ms, filePath, false)
		s.CounterInc("counter1", 1)
		s.GaugeSet("gauge1", 2.2)
		s.Save()

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"Gauge":{"gauge1":2.2},"Counter":{"counter1":1}}`,
			string(b),
		)
		s.Close()
	})
	os.Remove(filePath)

	filePath = fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("save sync", func(t *testing.T) {
		ms := memory.NewStorage()

		s := NewStorage(ms, filePath, true)
		s.CounterInc("counter1", 3)
		s.GaugeSet("gauge1", 4.4)

		b, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"Gauge":{"gauge1":4.4},"Counter":{"counter1":3}}`,
			string(b),
		)

		s.Close()
	})
	os.Remove(filePath)

	filePath = fmt.Sprintf("/tmp/random-%d.json", rand.Int())
	t.Run("multi load", func(t *testing.T) {
		ms := memory.NewStorage()
		ms.CounterInc("counter3", 3)
		ms.GaugeSet("gauge4", 4.4)

		s := NewStorage(ms, filePath, false)
		err := s.Save()
		require.NoError(t, err)

		ms2 := memory.NewStorage()
		s2 := NewStorage(ms2, filePath, false)
		err = s2.Load()
		require.NoError(t, err)

		cVal1, exists := ms.CounterGet("counter3")
		require.True(t, exists)
		cVal2, exists := ms2.CounterGet("counter3")
		require.True(t, exists)
		require.Equal(t, cVal1, cVal2)

		gVal1, exists := ms.GaugeGet("gauge4")
		require.True(t, exists)
		gVal2, exists := ms2.GaugeGet("gauge4")
		require.True(t, exists)
		require.Equal(t, gVal1, gVal2)

		// еще один load
		ms2.CounterInc("counter3", 9999)
		ms2.GaugeSet("gauge4", 9999)
		err = s2.Load()
		require.NoError(t, err)

		cVal2, exists = ms2.CounterGet("counter3")
		require.True(t, exists)
		require.Equal(t, int64(3), cVal2)
		gVal2, exists = ms2.GaugeGet("gauge4")
		require.True(t, exists)
		require.Equal(t, 4.4, gVal2)
	})
	os.Remove(filePath)
}

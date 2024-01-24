package fallback

import "github.com/gam6itko/go-musthave-metrics/internal/server/storage"

type Storage struct {
	inner    storage.Storage
	fallback storage.Storage
}

func NewStorage(inner storage.Storage, fallback storage.Storage) *Storage {
	return &Storage{
		inner,
		fallback,
	}
}

func (ths Storage) GaugeSet(name string, val float64) error {
	if err := ths.inner.GaugeSet(name, val); err == nil {
		return nil
	}
	return ths.fallback.GaugeSet(name, val)
}

func (ths Storage) GaugeGet(name string) (float64, error) {
	if result, err := ths.inner.GaugeGet(name); err == nil {
		return result, nil
	}
	return ths.fallback.GaugeGet(name)
}

func (ths Storage) GaugeAll() (map[string]float64, error) {
	if result, err := ths.inner.GaugeAll(); err == nil {
		return result, nil
	}
	return ths.fallback.GaugeAll()
}

func (ths Storage) CounterInc(name string, val int64) error {
	if err := ths.inner.CounterInc(name, val); err == nil {
		return nil
	}
	return ths.fallback.CounterInc(name, val)
}

func (ths Storage) CounterGet(name string) (int64, error) {
	if result, err := ths.inner.CounterGet(name); err == nil {
		return result, nil
	}
	return ths.fallback.CounterGet(name)
}

func (ths Storage) CounterAll() (map[string]int64, error) {
	if result, err := ths.inner.CounterAll(); err == nil {
		return result, nil
	}
	return ths.fallback.CounterAll()
}

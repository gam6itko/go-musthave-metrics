package memory

import (
	"context"
	"errors"
	"sync"
)

type Storage struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
	mux     sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		make(map[string]int64),
		make(map[string]float64),
		sync.RWMutex{},
	}
}

func (ths *Storage) CounterInc(_ context.Context, name string, val int64) error {
	ths.mux.Lock()
	defer ths.mux.Unlock()

	ths.Counter[name] += val
	return nil
}

func (ths *Storage) CounterGet(_ context.Context, name string) (int64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	if val, exists := ths.Counter[name]; exists {
		return val, nil
	}

	return 0, errors.New("not found")
}

func (ths *Storage) CounterAll(_ context.Context) (map[string]int64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	result := make(map[string]int64)
	for k, v := range ths.Counter {
		result[k] = v
	}

	return ths.Counter, nil
}

func (ths *Storage) GaugeSet(_ context.Context, name string, val float64) error {
	ths.mux.Lock()
	defer ths.mux.Unlock()

	ths.Gauge[name] = val
	return nil
}

func (ths *Storage) GaugeGet(_ context.Context, name string) (float64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	if val, ok := ths.Gauge[name]; ok {
		return val, nil
	}
	return 0.0, errors.New("not found")
}

func (ths *Storage) GaugeAll(_ context.Context) (map[string]float64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	result := make(map[string]float64)
	for k, v := range ths.Gauge {
		result[k] = v
	}

	return result, nil
}

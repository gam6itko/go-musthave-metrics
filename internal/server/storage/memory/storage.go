package memory

import (
	"context"
	"errors"
	"sync"
)

type Storage struct {
	counter map[string]int64
	gauge   map[string]float64
	mux     sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		make(map[string]int64),
		make(map[string]float64),
		sync.RWMutex{},
	}
}

func (ths *Storage) CounterInc(ctx context.Context, name string, val int64) error {
	ths.mux.Lock()
	defer ths.mux.Unlock()

	ths.counter[name] += val
	return nil
}

func (ths *Storage) CounterGet(ctx context.Context, name string) (int64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	if val, exists := ths.counter[name]; exists {
		return val, nil
	}

	return 0, errors.New("not found")
}

func (ths *Storage) CounterAll(ctx context.Context) (map[string]int64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	result := make(map[string]int64)
	for k, v := range ths.counter {
		result[k] = v
	}

	return ths.counter, nil
}

func (ths *Storage) GaugeSet(ctx context.Context, name string, val float64) error {
	ths.mux.Lock()
	defer ths.mux.Unlock()

	ths.gauge[name] = val
	return nil
}

func (ths *Storage) GaugeGet(ctx context.Context, name string) (float64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	if val, ok := ths.gauge[name]; ok {
		return val, nil
	}
	return 0.0, errors.New("not found")
}

func (ths *Storage) GaugeAll(ctx context.Context) (map[string]float64, error) {
	ths.mux.RLock()
	defer ths.mux.RUnlock()

	result := make(map[string]float64)
	for k, v := range ths.gauge {
		result[k] = v
	}

	return result, nil
}

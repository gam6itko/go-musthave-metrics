package memory

import (
	"context"
	"errors"
)

type Storage struct {
	Counter map[string]int64
	Gauge   map[string]float64
}

func NewStorage() *Storage {
	return &Storage{
		make(map[string]int64),
		make(map[string]float64),
	}
}

func (ths Storage) CounterInc(ctx context.Context, name string, val int64) error {
	ths.Counter[name] += val
	return nil
}

func (ths Storage) CounterGet(_ context.Context, name string) (int64, error) {
	if val, exists := ths.Counter[name]; exists {
		return val, nil
	}

	return 0, errors.New("not found")
}

func (ths Storage) CounterAll(_ context.Context) (map[string]int64, error) {
	return ths.Counter, nil
}

func (ths Storage) GaugeSet(_ context.Context, name string, val float64) error {
	ths.Gauge[name] = val
	return nil
}

func (ths Storage) GaugeGet(_ context.Context, name string) (float64, error) {
	if val, ok := ths.Gauge[name]; ok {
		return val, nil
	}
	return 0.0, errors.New("not found")
}

func (ths Storage) GaugeAll(_ context.Context) (map[string]float64, error) {
	return ths.Gauge, nil
}

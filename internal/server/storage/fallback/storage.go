package fallback

import (
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
)

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

func (ths Storage) GaugeSet(ctx context.Context, name string, val float64) error {
	if err := ths.inner.GaugeSet(ctx, name, val); err == nil {
		return nil
	}
	return ths.fallback.GaugeSet(ctx, name, val)
}

func (ths Storage) GaugeGet(ctx context.Context, name string) (float64, error) {
	if result, err := ths.inner.GaugeGet(ctx, name); err == nil {
		return result, nil
	}
	return ths.fallback.GaugeGet(ctx, name)
}

func (ths Storage) GaugeAll(ctx context.Context) (map[string]float64, error) {
	if result, err := ths.inner.GaugeAll(ctx); err == nil {
		return result, nil
	}
	return ths.fallback.GaugeAll(ctx)
}

func (ths Storage) CounterInc(ctx context.Context, name string, val int64) error {
	if err := ths.inner.CounterInc(ctx, name, val); err == nil {
		return nil
	}
	return ths.fallback.CounterInc(ctx, name, val)
}

func (ths Storage) CounterGet(ctx context.Context, name string) (int64, error) {
	if result, err := ths.inner.CounterGet(ctx, name); err == nil {
		return result, nil
	}
	return ths.fallback.CounterGet(ctx, name)
}

func (ths Storage) CounterAll(ctx context.Context) (map[string]int64, error) {
	if result, err := ths.inner.CounterAll(ctx); err == nil {
		return result, nil
	}
	return ths.fallback.CounterAll(ctx)
}

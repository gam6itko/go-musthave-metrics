package storage

import "context"

type Storage interface {
	GaugeSet(ctx context.Context, name string, val float64) error
	GaugeGet(ctx context.Context, name string) (float64, error)
	GaugeAll(ctx context.Context) (map[string]float64, error)

	CounterInc(ctx context.Context, name string, val int64) error
	CounterGet(ctx context.Context, name string) (int64, error)
	CounterAll(ctx context.Context) (map[string]int64, error)
}

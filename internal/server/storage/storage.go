package storage

import "context"

// IStorage хранит метрики отправленные агентом.
type IStorage interface {
	// GaugeSet установить значение val шкале с именем name.
	GaugeSet(ctx context.Context, name string, val float64) error
	GaugeGet(ctx context.Context, name string) (float64, error)
	GaugeAll(ctx context.Context) (map[string]float64, error)

	// CounterInc увеличивает счётцик c именем name на значение val.
	CounterInc(ctx context.Context, name string, val int64) error
	CounterGet(ctx context.Context, name string) (int64, error)
	// CounterAll векрнёт все имеющиеся счётчики.
	CounterAll(ctx context.Context) (map[string]int64, error)
}

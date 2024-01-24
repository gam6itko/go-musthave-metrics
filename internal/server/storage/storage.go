package storage

type Storage interface {
	GaugeSet(name string, val float64) error
	GaugeGet(name string) (float64, error)
	GaugeAll() (map[string]float64, error)

	CounterInc(name string, val int64) error
	CounterGet(name string) (int64, error)
	CounterAll() (map[string]int64, error)
}

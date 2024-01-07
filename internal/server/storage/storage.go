package storage

type IMetricStorage interface {
	GaugeSet(name string, val float64)
	GaugeGet(name string) (float64, bool)
	GaugeAll() map[string]float64

	CounterInc(name string, val int64)
	CounterGet(name string) (int64, bool)
	CounterAll() map[string]int64
}

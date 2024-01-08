package memory

type Storage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewStorage() *Storage {
	return &Storage{
		make(map[string]float64),
		make(map[string]int64),
	}
}

func (ths Storage) GaugeSet(name string, val float64) {
	ths.Gauge[name] = val
}

func (ths Storage) GaugeGet(name string) (float64, bool) {
	val, ok := ths.Gauge[name]
	return val, ok
}

func (ths Storage) GaugeAll() map[string]float64 {
	return ths.Gauge
}

func (ths Storage) CounterInc(name string, val int64) {
	ths.Counter[name] += val
}

func (ths Storage) CounterGet(name string) (int64, bool) {
	if val, exists := ths.Counter[name]; exists {
		return val, true
	}

	return 0, false
}

func (ths Storage) CounterAll() map[string]int64 {
	return ths.Counter
}

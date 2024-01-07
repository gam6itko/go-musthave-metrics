package memory

type Storage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewStorage() Storage {
	return Storage{
		make(map[string]float64),
		make(map[string]int64),
	}
}

func (ths Storage) GaugeSet(name string, val float64) {
	ths.gauge[name] = val
}

func (ths Storage) GaugeGet(name string) (float64, bool) {
	val, ok := ths.gauge[name]
	return val, ok
}

func (ths Storage) GaugeAll() map[string]float64 {
	return ths.gauge
}

func (ths Storage) CounterInc(name string, val int64) {
	ths.counter[name] += val
}

func (ths Storage) CounterGet(name string) (int64, bool) {
	if val, exists := ths.counter[name]; exists {
		return val, true
	}

	return 0, false
}

func (ths Storage) CounterAll() map[string]int64 {
	return ths.counter
}

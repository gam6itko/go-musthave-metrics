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

func (s Storage) GaugeSet(name string, val float64) {
	s.gauge[name] = val
}

func (s Storage) GaugeGet(name string) (float64, bool) {
	val, ok := s.gauge[name]
	return val, ok
}

func (s Storage) GaugeAll() map[string]float64 {
	return s.gauge
}

func (s Storage) CounterInc(name string, val int64) {
	s.counter[name] += val
}

func (s Storage) CounterGet(name string) (int64, bool) {
	if val, exists := s.counter[name]; exists {
		return val, true
	}

	return 0, false
}

func (s Storage) CounterAll() map[string]int64 {
	return s.counter
}

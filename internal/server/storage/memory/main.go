package memory

var DefaultStorage Storage

func init() {
	DefaultStorage = NewStorage()
}

func GaugeSet(name string, val float64) {
	DefaultStorage.gauge[name] = val
}

func GaugeGet(name string) (float64, bool) {
	val, ok := DefaultStorage.gauge[name]
	return val, ok
}

func GaugeAll() map[string]float64 {
	return DefaultStorage.gauge
}

func CounterInc(name string, val int64) {
	DefaultStorage.counter[name] += val
}

func CounterGet(name string) (int64, bool) {
	if val, exists := DefaultStorage.counter[name]; exists {
		return val, true
	}

	return 0, false
}

func CounterAll() map[string]int64 {
	return DefaultStorage.counter
}

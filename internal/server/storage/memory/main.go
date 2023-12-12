package memory

type memStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var storage memStorage

func init() {
	storage = memStorage{
		make(map[string]float64),
		make(map[string]int64),
	}
}

func GaugeSet(name string, val float64) {
	storage.gauge[name] = val
}

func GaugeGet(name string) (float64, bool) {
	val, ok := storage.gauge[name]
	return val, ok
}

func GaugeAll() map[string]float64 {
	return storage.gauge
}

func CounterInc(name string, val int64) {
	storage.counter[name] += val
}

func CounterGet(name string) (int64, bool) {
	if val, exists := storage.counter[name]; exists {
		return val, true
	}

	return 0, false
}

func CounterAll() map[string]int64 {
	return storage.counter
}

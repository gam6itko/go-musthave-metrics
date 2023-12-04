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

func CounterInc(name string, val int64) {
	if _, exists := storage.counter[name]; !exists {
		storage.counter[name] = 0
	}
	storage.counter[name] += val
}

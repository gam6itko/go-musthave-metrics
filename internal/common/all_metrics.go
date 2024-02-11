package common

type AllMetrics struct {
	Counter map[string]int64
	Gauge   map[string]float64
}

func NewAllMetrics(counter map[string]int64, gauge map[string]float64) AllMetrics {
	return AllMetrics{
		counter,
		gauge,
	}
}

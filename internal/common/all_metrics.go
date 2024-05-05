package common

type AllMetrics struct {
	Counter map[string]int64   `json:"counter"`
	Gauge   map[string]float64 `json:"gauge"`
}

func NewAllMetrics(counter map[string]int64, gauge map[string]float64) AllMetrics {
	return AllMetrics{
		counter,
		gauge,
	}
}

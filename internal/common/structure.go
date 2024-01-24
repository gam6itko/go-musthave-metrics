package common

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge

	//todo за это стыдно, но я не знаю как иначе сделать Delta и Value ссылкой. Мне за это стыдно.
	DeltaForRef int64
	ValueForRef float64
}

package common

// Metrics отдаваемые и принимаемые метрики.
// nil используется чтобы в json не было delta=0 если у нас gauge и наоборот.
type Metrics struct {
	// Delta значение метрики в случае передачи counter
	Delta *int64 `json:"delta,omitempty"`
	// Value значение метрики в случае передачи gauge
	Value *float64 `json:"value,omitempty"`

	// ID содержит имя метрики
	ID string `json:"id"`
	// MType принимает значение 'gauge' или 'counter'
	MType string `json:"type"`

	//todo я не знаю как иначе сделать Delta и Value ссылкой. Мне за это стыдно.
	DeltaForRef int64
	ValueForRef float64
}

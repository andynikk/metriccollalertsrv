package encoding

import (
	"encoding/json"
)

type ArrMetrics []Metrics

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m *Metrics) MarshalMetrica() (val []byte, err error) {
	arrJSON, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return nil, err
	}

	return arrJSON, nil
}

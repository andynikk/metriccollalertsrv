package encoding

import (
	"encoding/json"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  []byte   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m Metrics) MarshalMetrica() (val []byte, err error) {
	var arrJSON, errMarshal = json.Marshal(m)
	if errMarshal != nil {
		var bt []byte
		return bt, errMarshal
	}

	return arrJSON, nil
}

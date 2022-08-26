package encoding

import (
	"encoding/json"
)

//type MetricsGauge = map[string]repository.Gauge

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"mtype"`           // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metrics) MarshalMetrica() (val []byte, err error) {
	var arrJSON, errMarshal = json.Marshal(m)
	if errMarshal != nil {
		var bt []byte
		return bt, errMarshal
	}

	return arrJSON, nil
}

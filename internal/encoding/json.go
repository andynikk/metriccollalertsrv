package encoding

import (
	"encoding/json"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricsGauge = map[string]repository.Gauge

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metrics) MarshalMetrica() (val []byte, err error) {
	var arrJson, errMarshal = json.Marshal(m)
	if errMarshal != nil {
		var bt []byte
		return bt, errMarshal
	}

	return arrJson, nil
}

func (m Metrics) Gauge() repository.Gauge {
	return repository.Gauge(*m.Value)
}

func (m Metrics) Counter() repository.Counter {
	return repository.Counter(*m.Delta)
}

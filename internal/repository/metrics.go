package repository

import (
	"fmt"
	"strings"
)

type Gauge float64
type Counter int64

type MetricsType = map[string]interface{}
type MapMetrics = map[string]MetricsType

var Metrics = make(MapMetrics)

type Metric interface {
	SetVal(string, string) error
	String() string
	Type() string
	GetVal(string) string
}

func (g Gauge) SetVal(mapa MetricsType, nameMetric string) {
	mapa[nameMetric] = g
}

func (c Counter) SetVal(mapa MetricsType, nameMetric string) {

	if _, findKey := mapa[nameMetric]; !findKey {
		mapa[nameMetric] = c
	} else {
		predVal := mapa[nameMetric].(Counter)
		val := int64(predVal) + int64(c)

		mapa[nameMetric] = Counter(val)
	}

}

func (g Gauge) String() string {
	fg := float64(g)
	return fmt.Sprintf("%g", fg)
}

func (c Counter) String() string {
	return fmt.Sprintf("%d", int64(c))
}

func (g Gauge) Type() string {
	return "gauge"
}

func (c Counter) Type() string {
	return "counter"
}

func TextMetricsAndValue() string {
	const msgFormat = "%s = %s"

	var msg []string
	for _, mapa := range Metrics {
		for key, val := range mapa {
			msg = append(msg, fmt.Sprintf(msgFormat, key, val))
		}
	}

	return strings.Join(msg, "\n")
}

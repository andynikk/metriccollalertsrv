package repository

import (
	"fmt"
	"sync"
)

type Gauge float64
type Counter int64

type MutexCounters struct {
	mx sync.Mutex
	m  MetricsType
}

type MetricsType = map[string]interface{}
type MapMetrics = map[string]MetricsType

type Metric interface {
	SetVal(string, string) error
	String() string
	Type() string
	GetVal(string) string
}

func (g Gauge) SetVal(mapa MetricsType, nameMetric string) {
	var lock sync.Mutex
	defer lock.Unlock()

	lock.Lock()
	mapa[nameMetric] = g
}

func (c Counter) SetVal(mapa MetricsType, nameMetric string) {
	var lock sync.Mutex
	defer lock.Unlock()

	lock.Lock()

	if _, findKey := mapa[nameMetric]; !findKey {
		mapa[nameMetric] = c

	} else {

		mapa[nameMetric] = mapa[nameMetric].(Counter) + c
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

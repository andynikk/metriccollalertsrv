package repository

import (
	"fmt"
)

type Gauge float64
type Counter int64

type MetricsType = map[string]interface{}
type MapMetrics = map[string]MetricsType

//type RepStore struct {
//	MutexRepo MapMetrics
//	Router    chi.Router
//	mx        sync.Mutex
//}

type Metric interface {
	String() string
	Type() string
	GetVal(string) string
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

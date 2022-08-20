package repository

import (
	"fmt"
)

type Gauge float64
type Counter int64

type MetricsType = map[string]interface{}
type MapMetrics = map[string]MetricsType

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

func (g Gauge) Float64() float64 {
	return float64(g)
}

func (c Counter) Type() string {
	return "counter"
}

func (c Counter) Int64() int64 {
	return int64(c)
}

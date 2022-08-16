package repository

import (
	"fmt"
	"sync"
)

type Gauge float64
type Counter int64

type MetricsType = map[string]interface{}
type MapMetrics = map[string]MutexTypeMetrics

type MutexTypeMetrics struct {
	M  MetricsType
	mx sync.Mutex
}

type Metric interface {
	SetVal(string, string) error
	String() string
	Type() string
	GetVal(string) string
}

func (c *MutexTypeMetrics) SetCounter(key string, value Counter) {
	c.mx.Lock()

	if _, findKey := c.M[key]; !findKey {
		c.M[key] = value

	} else {
		c.M[key] = c.M[key].(Counter) + value
	}

	c.mx.Unlock()
}

func (c *MutexTypeMetrics) SetGauge(key string, value Gauge) {
	c.mx.Lock()

	c.M[key] = value

	c.mx.Unlock()
}

func (c *MutexTypeMetrics) ValueCounter(key string, val Counter) Counter {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.M[key].(Counter)
}

func (c *MutexTypeMetrics) ValueGauge(key string, val Gauge) Gauge {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.M[key].(Gauge)
}

func (g Gauge) SetVal(mapa MutexTypeMetrics, nameMetric string) {

	mapa.M[nameMetric] = g
}

func (c Counter) SetVal(mapa MutexTypeMetrics, nameMetric string) {
	var lock sync.Mutex
	defer lock.Unlock()

	lock.Lock()

	if _, findKey := mapa.M[nameMetric]; !findKey {
		mapa.M[nameMetric] = c

	} else {

		mapa.M[nameMetric] = mapa.M[nameMetric].(Counter) + c
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

package repository

import (
	"fmt"
	"strconv"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

type Gauge float64
type Counter int64

type MapMetrics = map[string]Metric

type Metric interface {
	String() string
	Type() string
	Set(v encoding.Metrics)
	SetFromText(metValue string) bool
	GetMetrics(id string, mType string) encoding.Metrics
}

func (g *Gauge) String() string {

	return fmt.Sprintf("%g", *g)
}

func (g *Gauge) Type() string {
	return "gauge"
}

func (g *Gauge) GetMetrics(mType string, id string) encoding.Metrics {

	value := float64(*g)
	mt := encoding.Metrics{ID: id, MType: mType, Value: &value}

	return mt
}

func (g *Gauge) Set(v encoding.Metrics) {

	*g = Gauge(*v.Value)

}

func (g *Gauge) SetFromText(metValue string) bool {

	predVal, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		fmt.Println("error convert type")
		return false
	}
	*g = Gauge(predVal)

	return true

}

///////////////////////////////////////////////////////////////////////////////

func (c *Counter) Set(v encoding.Metrics) {

	*c = *c + Counter(*v.Delta)
}

func (c *Counter) SetFromText(metValue string) bool {

	predVal, err := strconv.ParseInt(metValue, 10, 64)
	if err != nil {
		fmt.Println("error convert type")
		return false
	}
	*c = *c + Counter(predVal)

	return true

}

func (c *Counter) String() string {

	return fmt.Sprintf("%d", *c)
}

func (c *Counter) GetMetrics(mType string, id string) encoding.Metrics {

	delta := int64(*c)
	mt := encoding.Metrics{ID: id, MType: mType, Delta: &delta}

	return mt
}

func (c *Counter) Type() string {
	return "counter"
}

////////////////////////////////////////////////////////////////////////////////

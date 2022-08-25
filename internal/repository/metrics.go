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
	GetMetrics(id string, mType string) encoding.Metrics
	Set(v encoding.Metrics)
	Float64() float64
	Int64() int64
	SetFromText(metValue string) int
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

func (g *Gauge) SetFromText(metValue string) int {

	predVal, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		fmt.Println("error convert type")
		return 400
	}
	*g = Gauge(predVal)

	return 200

}

func (g *Gauge) Int64() int64 {
	return int64(*g)
}

func (g *Gauge) Float64() float64 {
	return float64(*g)
}

///////////////////////////////////////////////////////////////////////////////

func (c *Counter) Set(v encoding.Metrics) {
	//ival, ok := val.(int64)
	//if ok {
	//	*c = *c + Counter(ival)
	//}
	*c = *c + Counter(*v.Delta)
}

func (c *Counter) SetFromText(metValue string) int {

	predVal, err := strconv.ParseInt(metValue, 10, 64)
	if err != nil {
		fmt.Println("error convert type")
		return 400
	}
	*c = *c + Counter(predVal)

	return 200

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

func (c *Counter) Int64() int64 {
	return int64(*c)
}

func (c *Counter) Float64() float64 {
	return float64(*c)
}

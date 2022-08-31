package repository

import (
	"fmt"
	"net/http"
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
	SetFromText(metValue string) int
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

func (g *Gauge) SetFromText(metValue string) int {

	predVal, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		fmt.Println("error convert type")
		return http.StatusBadRequest
	}
	*g = Gauge(predVal)

	return http.StatusOK

}

///////////////////////////////////////////////////////////////////////////////

func (c *Counter) Set(v encoding.Metrics) {

	*c = *c + Counter(*v.Delta)
}

func (c *Counter) SetFromText(metValue string) int {

	predVal, err := strconv.ParseInt(metValue, 10, 64)
	if err != nil {
		fmt.Println("error convert type")
		return http.StatusBadRequest
	}
	*c = *c + Counter(predVal)

	return http.StatusOK

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

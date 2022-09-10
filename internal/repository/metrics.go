package repository

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
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
	GetMetrics(id string, mType string, hashKey string) encoding.Metrics
}

func (g *Gauge) String() string {

	return fmt.Sprintf("%g", *g)
}

func (g *Gauge) Type() string {
	return "gauge"
}

func (g *Gauge) GetMetrics(mType string, id string, hashKey string) encoding.Metrics {

	value := float64(*g)
	msg := fmt.Sprintf("%s:%s:%f", id, mType, value)
	heshVal := cryptohash.HeshSHA256(msg, hashKey)

	mt := encoding.Metrics{ID: id, MType: mType, Value: &value, Hash: heshVal}

	return mt
}

func (g *Gauge) Set(v encoding.Metrics) {

	*g = Gauge(*v.Value)

}

func (g *Gauge) SetFromText(metValue string) bool {

	predVal, err := strconv.ParseFloat(metValue, 64)
	if err != nil {
		constants.Logger.Error().Err(errors.New("error convert type"))

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
		constants.Logger.Error().Err(errors.New("error convert type"))

		return false
	}
	*c = *c + Counter(predVal)

	return true

}

func (c *Counter) String() string {

	return fmt.Sprintf("%d", *c)
}

func (c *Counter) GetMetrics(mType string, id string, hashKey string) encoding.Metrics {

	delta := int64(*c)

	msg := fmt.Sprintf("%s:%s:%d", id, mType, delta)
	heshVal := cryptohash.HeshSHA256(msg, hashKey)

	mt := encoding.Metrics{ID: id, MType: mType, Delta: &delta, Hash: heshVal}

	return mt
}

func (c *Counter) Type() string {
	return "counter"
}

////////////////////////////////////////////////////////////////////////////////

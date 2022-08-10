package repository

import (
	"errors"
	"fmt"
	"strconv"
)

var typeMetGauge = map[string]Gauge{}
var typeMetCounter = map[string]Counter{}

type Gauge float64
type Counter int64

var Metrics = map[string]interface{}{}

type Metric interface {
	SetVal(string, string) error
	String() string
	Type() string
	GetVal(string) string
}

func (g Gauge) SetVal(nameMetric string, valMetric string) error {

	val, err := strconv.ParseFloat(valMetric, 64)
	if err != nil {
		return errors.New("error convert type")
	}

	Metrics[nameMetric] = Gauge(val)
	return nil
}

func (c Counter) SetVal(nameMetric string, valMetric string) error {

	var val, err = strconv.ParseInt(valMetric, 10, 64)
	if err != nil {
		return errors.New("error convert type counter")
	}

	if _, findKey := Metrics[nameMetric]; findKey {
		predVal := Metrics[nameMetric].(Counter)
		Metrics[nameMetric] = Counter(val + int64(predVal))
	} else {
		Metrics[nameMetric] = Counter(val)
	}

	return nil
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

func (g Gauge) GetVal(nameMetric string) string {
	return nameMetric //Metrics[nameMetric].(Gauge)
}

func (c Counter) GetVal(nameMetric string) string {
	return nameMetric //Metrics[nameMetric].(Counter)
}

func returnMetric(typeMetric string) (Metric, error) {
	var mtc Metric

	if _, findKey := typeMetGauge[typeMetric]; findKey {
		mtc = typeMetGauge[typeMetric]
		return mtc, nil
	}
	if _, findKey := typeMetCounter[typeMetric]; findKey {
		mtc = typeMetCounter[typeMetric]
		return mtc, nil
	}

	return nil, errors.New("501")
}

func SetValue(typeMetric string, nameMetric string, valMetric string) error {
	mtc, err := returnMetric(typeMetric)
	if err != nil {
		return err
	}

	err = mtc.SetVal(nameMetric, valMetric)
	if err != nil {
		return errors.New("400")
	}

	return nil
}

func StringValue(typeMetric string, nameMetric string) string {

	var mtc Metric

	if _, findKey := typeMetGauge[typeMetric]; findKey {
		mtc = Metrics[nameMetric].(Gauge)
	} else {
		mtc = Metrics[nameMetric].(Counter)
	}

	return mtc.String()

}

func GetValue(typeMetric string, nameMetric string) string {

	var mtc Metric

	if _, findKey := typeMetGauge[typeMetric]; findKey {
		mtc = Metrics[nameMetric].(Gauge)
	} else {
		mtc = Metrics[nameMetric].(Counter)
	}

	return mtc.GetVal(nameMetric)

}

func RefTypeMepStruc() {
	typeMetGauge["gauge"] = Gauge(0)
	typeMetCounter["counter"] = Counter(0)
}

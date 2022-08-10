package repository

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

var typeMetGauge = map[string]Gauge{}
var typeMetCounter = map[string]Counter{}

type Gauge float64
type Counter int64

var Metrics = map[string]interface{}{}

type Metric interface {
	SetVal(string, string, string) error
	String() string
	//GetVal(string) (Gauge, Counter)
}

func (g Gauge) SetVal(typeMetric string, nameMetric string, valMetric string) error {
	if typeMetric != "gauge" {
		return errors.New("This not guage")
	}
	val, err := strconv.ParseFloat(valMetric, 64)
	if err != nil {
		fmt.Println(err)
		return errors.New("error convert type")
	}

	Metrics[nameMetric] = Gauge(val)
	return nil
}

func (c Counter) SetVal(typeMetric string, nameMetric string, valMetric string) error {
	if typeMetric != "counter" {
		return errors.New("this not counter")
	}

	var val, err = strconv.ParseInt(valMetric, 10, 64)
	if err != nil {
		return errors.New("error convert type counter")
	}

	if _, findKey := Metrics[nameMetric]; findKey {
		predVal, findKey := Metrics[nameMetric].(Counter)
		if !findKey {
			log.Fatal("could not assert value to int")
		}
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

func (g Gauge) GetVal(nameMetric string) Gauge {
	return Metrics[nameMetric].(Gauge)
}

func (c Counter) GetVal(nameMetric string) Counter {
	return Metrics[nameMetric].(Counter)
}

func SetValue(typeMetric string, nameMetric string, valMetric string) {
	var mtc Metric

	if _, findKey := typeMetGauge[typeMetric]; findKey {
		mtc = typeMetGauge[typeMetric]
	} else {
		mtc = typeMetCounter[typeMetric]
	}

	err := mtc.SetVal(typeMetric, nameMetric, valMetric)
	if err != nil {
		log.Fatal("The value is not set")
	}
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

func GetValue(typeMetric string, nameMetric string) Metric {
	var mtc Metric

	if _, findKey := typeMetGauge[typeMetric]; findKey {
		mtc = Metrics[nameMetric].(Gauge)
	} else {
		mtc = Metrics[nameMetric].(Counter)
	}

	return mtc

}

func RefTypeMepStruc() {
	typeMetGauge["gauge"] = Gauge(0)
	typeMetCounter["counter"] = Counter(0)
}

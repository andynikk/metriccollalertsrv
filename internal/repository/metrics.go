package repository

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

type Gauge float64
type Counter int64

type MetricType int

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

const (
	GaugeMetric MetricType = iota
	CounterMetric
)

type MapMetrics = map[string]Metric

type StoreMetrics struct {
	MapTypeStore  MapTypeStore
	StoreInterval time.Duration
	HashKey       string
	MX            *sync.Mutex
	Repo          MapMetrics
}

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
		constants.Logger.ErrorLog(errors.New("error convert type"))

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
		constants.Logger.ErrorLog(errors.New("error convert type"))

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

//func (sm *StoreMetrics) BackupData(ctx context.Context, cancelFunc context.CancelFunc) {

func (sm *StoreMetrics) BackupData() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	saveTicker := time.NewTicker(sm.StoreInterval)
	for {
		select {
		case <-saveTicker.C:
			for _, val := range sm.MapTypeStore {
				val.WriteMetric(sm.PrepareDataBU())
			}
		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (sm *StoreMetrics) PrepareDataBU() encoding.ArrMetrics {

	var storedData encoding.ArrMetrics
	for key, val := range sm.Repo {
		storedData = append(storedData, val.GetMetrics(val.Type(), key, sm.HashKey))
	}
	return storedData
}

func (sm *StoreMetrics) RestoreData() {
	for _, val := range sm.MapTypeStore {
		arrMetrics, err := val.GetMetric()
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}

		sm.MX.Lock()
		defer sm.MX.Unlock()

		for _, val := range arrMetrics {
			sm.SetValueInMapJSON(val)
		}
	}
}

func (sm *StoreMetrics) SetValueInMapJSON(v encoding.Metrics) int {

	var heshVal string

	switch v.MType {
	case GaugeMetric.String():
		var valValue float64
		valValue = *v.Value

		msg := fmt.Sprintf("%s:gauge:%f", v.ID, valValue)
		heshVal = cryptohash.HeshSHA256(msg, sm.HashKey)
		if _, findKey := sm.Repo[v.ID]; !findKey {
			valG := Gauge(0)
			sm.Repo[v.ID] = &valG
		}
	case CounterMetric.String():
		var valDelta int64
		valDelta = *v.Delta

		msg := fmt.Sprintf("%s:counter:%d", v.ID, valDelta)
		heshVal = cryptohash.HeshSHA256(msg, sm.HashKey)
		if _, findKey := sm.Repo[v.ID]; !findKey {
			valC := Counter(0)
			sm.Repo[v.ID] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	heshAgent := []byte(v.Hash)
	heshServer := []byte(heshVal)

	hmacEqual := hmac.Equal(heshServer, heshAgent)

	constants.Logger.InfoLog(fmt.Sprintf("-- %s - %s", v.Hash, heshVal))

	if v.Hash != "" && !hmacEqual {
		constants.Logger.InfoLog(fmt.Sprintf("++ %s - %s", v.Hash, heshVal))
		return http.StatusBadRequest
	}
	constants.Logger.InfoLog(fmt.Sprintf("** %s %s %v %d", v.ID, v.MType, v.Value, v.Delta))

	sm.Repo[v.ID].Set(v)
	return http.StatusOK

}

func (sm *StoreMetrics) SetValueInMap(metType string, metName string, metValue string) int {

	switch metType {
	case GaugeMetric.String():
		if val, findKey := sm.Repo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valG := Gauge(0)
			if ok := valG.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			sm.Repo[metName] = &valG
		}

	case CounterMetric.String():
		if val, findKey := sm.Repo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valC := Counter(0)
			if ok := valC.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			sm.Repo[metName] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	return http.StatusOK
}

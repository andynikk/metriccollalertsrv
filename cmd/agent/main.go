package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricsGauge map[string]repository.Gauge
type emtyArrMetrics []encoding.Metrics

type agent struct {
	MetricsGauge MetricsGauge
	PollCount    int64
	Cfg          environment.AgentConfig
}

func (a *agent) fillMetric(mem *runtime.MemStats) {

	a.MetricsGauge["Alloc"] = repository.Gauge(mem.Alloc)
	a.MetricsGauge["BuckHashSys"] = repository.Gauge(mem.BuckHashSys)
	a.MetricsGauge["Frees"] = repository.Gauge(mem.Frees)
	a.MetricsGauge["GCCPUFraction"] = repository.Gauge(mem.GCCPUFraction)
	a.MetricsGauge["GCSys"] = repository.Gauge(mem.GCSys)
	a.MetricsGauge["HeapAlloc"] = repository.Gauge(mem.HeapAlloc)
	a.MetricsGauge["HeapIdle"] = repository.Gauge(mem.HeapIdle)
	a.MetricsGauge["HeapInuse"] = repository.Gauge(mem.HeapInuse)
	a.MetricsGauge["HeapObjects"] = repository.Gauge(mem.HeapObjects)
	a.MetricsGauge["HeapReleased"] = repository.Gauge(mem.HeapReleased)
	a.MetricsGauge["HeapSys"] = repository.Gauge(mem.HeapSys)
	a.MetricsGauge["LastGC"] = repository.Gauge(mem.LastGC)
	a.MetricsGauge["Lookups"] = repository.Gauge(mem.Lookups)
	a.MetricsGauge["MCacheInuse"] = repository.Gauge(mem.MCacheInuse)
	a.MetricsGauge["MCacheSys"] = repository.Gauge(mem.MCacheSys)
	a.MetricsGauge["MSpanInuse"] = repository.Gauge(mem.MSpanInuse)
	a.MetricsGauge["MSpanSys"] = repository.Gauge(mem.MSpanSys)
	a.MetricsGauge["Mallocs"] = repository.Gauge(mem.Mallocs)
	a.MetricsGauge["NextGC"] = repository.Gauge(mem.NextGC)
	a.MetricsGauge["NumForcedGC"] = repository.Gauge(mem.NumForcedGC)
	a.MetricsGauge["NumGC"] = repository.Gauge(mem.NumGC)
	a.MetricsGauge["OtherSys"] = repository.Gauge(mem.OtherSys)
	a.MetricsGauge["PauseTotalNs"] = repository.Gauge(mem.PauseTotalNs)
	a.MetricsGauge["StackInuse"] = repository.Gauge(mem.StackInuse)
	a.MetricsGauge["StackSys"] = repository.Gauge(mem.StackSys)
	a.MetricsGauge["Sys"] = repository.Gauge(mem.Sys)
	a.MetricsGauge["TotalAlloc"] = repository.Gauge(mem.TotalAlloc)
	a.MetricsGauge["RandomValue"] = repository.Gauge(rand.Float64())

	a.PollCount = a.PollCount + 1

}

func (a *agent) metrixScan() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	a.fillMetric(&mem)
}

func (a *agent) Post2Server(allMterics *[]byte) error {

	addressPost := fmt.Sprintf("http://%s/updates", a.Cfg.Address)
	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(*allMterics))
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return errors.New("-- ошибка отправки данных на сервер (1)")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return errors.New("-- ошибка отправки данных на сервер (2)")
	}
	defer resp.Body.Close()

	return nil
}

func prepareMetrics(allMetrics *emtyArrMetrics) ([]byte, error) {

	arrMetrics, err := json.MarshalIndent(&allMetrics, "", " ")
	if err != nil {
		return nil, err
	}

	gziparrMetrics, err := compression.Compress(arrMetrics)
	if err != nil {
		return nil, err
	}
	return gziparrMetrics, nil
}

func (a *agent) MakeRequest() {

	allMetrics := make(emtyArrMetrics, 0)

	butch := 10

	sch := 0
	for key, val := range a.MetricsGauge {
		valFloat64 := float64(val)

		msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
		heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)

		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
		allMetrics = append(allMetrics, metrica)
		sch++
		if sch == butch {
			sch = 0
			if len(allMetrics) != 0 {
				gziparrMterica, err := prepareMetrics(&allMetrics)
				if err != nil {
					constants.InfoLevel.Error().Err(err)
				}
				if err := a.Post2Server(&gziparrMterica); err != nil {
					constants.InfoLevel.Error().Err(err)
				}
				allMetrics = make(emtyArrMetrics, 0)
			}
		}
	}

	cPollCount := repository.Counter(a.PollCount)
	msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.PollCount)
	heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)
	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.PollCount, Hash: heshVal}
	allMetrics = append(allMetrics, metrica)

	gziparrMetrics, err := prepareMetrics(&allMetrics)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
	}
	if err := a.Post2Server(&gziparrMetrics); err != nil {
		constants.InfoLevel.Error().Err(err)
	}
}

func main() {

	agent := agent{
		Cfg:          environment.SetConfigAgent(),
		MetricsGauge: make(MetricsGauge),
		PollCount:    0,
	}

	updateTicker := time.NewTicker(agent.Cfg.PollInterval)
	reportTicker := time.NewTicker(agent.Cfg.ReportInterval)

	for {
		select {
		case <-updateTicker.C:
			agent.metrixScan()
		case <-reportTicker.C:
			agent.MakeRequest()
		}
	}

}

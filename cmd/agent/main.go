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

func (a *agent) Post2Server(arrMterica *[]byte) error {

	addressPost := fmt.Sprintf("http://%s/updates", a.Cfg.Address)
	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(*arrMterica))
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

func (a *agent) MakeRequest() {

	var allMterics []interface{}

	for key, val := range a.MetricsGauge {
		valFloat64 := float64(val)

		msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
		heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)

		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
		allMterics = append(allMterics, metrica)
	}

	cPollCount := repository.Counter(a.PollCount)

	msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.PollCount)
	heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)

	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.PollCount, Hash: heshVal}
	allMterics = append(allMterics, metrica)

	arrMterics, err := json.MarshalIndent(allMterics, "", " ")
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	gziparrMterica, err := compression.Compress(arrMterics)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}
	if err := a.Post2Server(&gziparrMterica); err != nil {
		constants.InfoLevel.Error().Err(err)
		return
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

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
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
	PollCount    int64
	Cfg          environment.AgentConfig
	MX           sync.Mutex
	MetricsGauge MetricsGauge
}

func (a *agent) fillMetric(mems *runtime.MemStats) {

	a.MX.Lock()
	defer a.MX.Unlock()

	a.MetricsGauge["Alloc"] = repository.Gauge(mems.Alloc)
	a.MetricsGauge["BuckHashSys"] = repository.Gauge(mems.BuckHashSys)
	a.MetricsGauge["Frees"] = repository.Gauge(mems.Frees)
	a.MetricsGauge["GCCPUFraction"] = repository.Gauge(mems.GCCPUFraction)
	a.MetricsGauge["GCSys"] = repository.Gauge(mems.GCSys)
	a.MetricsGauge["HeapAlloc"] = repository.Gauge(mems.HeapAlloc)
	a.MetricsGauge["HeapIdle"] = repository.Gauge(mems.HeapIdle)
	a.MetricsGauge["HeapInuse"] = repository.Gauge(mems.HeapInuse)
	a.MetricsGauge["HeapObjects"] = repository.Gauge(mems.HeapObjects)
	a.MetricsGauge["HeapReleased"] = repository.Gauge(mems.HeapReleased)
	a.MetricsGauge["HeapSys"] = repository.Gauge(mems.HeapSys)
	a.MetricsGauge["LastGC"] = repository.Gauge(mems.LastGC)
	a.MetricsGauge["Lookups"] = repository.Gauge(mems.Lookups)
	a.MetricsGauge["MCacheInuse"] = repository.Gauge(mems.MCacheInuse)
	a.MetricsGauge["MCacheSys"] = repository.Gauge(mems.MCacheSys)
	a.MetricsGauge["MSpanInuse"] = repository.Gauge(mems.MSpanInuse)
	a.MetricsGauge["MSpanSys"] = repository.Gauge(mems.MSpanSys)
	a.MetricsGauge["Mallocs"] = repository.Gauge(mems.Mallocs)
	a.MetricsGauge["NextGC"] = repository.Gauge(mems.NextGC)
	a.MetricsGauge["NumForcedGC"] = repository.Gauge(mems.NumForcedGC)
	a.MetricsGauge["NumGC"] = repository.Gauge(mems.NumGC)
	a.MetricsGauge["OtherSys"] = repository.Gauge(mems.OtherSys)
	a.MetricsGauge["PauseTotalNs"] = repository.Gauge(mems.PauseTotalNs)
	a.MetricsGauge["StackInuse"] = repository.Gauge(mems.StackInuse)
	a.MetricsGauge["StackSys"] = repository.Gauge(mems.StackSys)
	a.MetricsGauge["Sys"] = repository.Gauge(mems.Sys)
	a.MetricsGauge["TotalAlloc"] = repository.Gauge(mems.TotalAlloc)
	a.MetricsGauge["RandomValue"] = repository.Gauge(rand.Float64())

	a.PollCount = a.PollCount + 1

}

func (a *agent) metrixOtherScan() {

	a.MX.Lock()
	defer a.MX.Unlock()

	swapMemory, err := mem.SwapMemory()
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	a.MetricsGauge["TotalMemory"] = repository.Gauge(swapMemory.Total)
	a.MetricsGauge["FreeMemory"] = repository.Gauge(swapMemory.Free)

	cpuUtilization, _ := cpu.Percent(1*time.Second, false)
	for _, val := range cpuUtilization {
		a.MetricsGauge["CPUutilization1"] = repository.Gauge(val)
		break
	}
}

func (a *agent) metrixScan() {
	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)
	a.fillMetric(&mems)
}

func (a *agent) Post2Server(allMterics *[]byte) error {

	addressPost := fmt.Sprintf("http://%s/updates", a.Cfg.Address)
	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(*allMterics))
	if err != nil {

		constants.Logger.ErrorLog(err)

		return errors.New("-- ошибка отправки данных на сервер (1)")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
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
	chanMatrics := make(chan encoding.Metrics, constants.ButchSize)

	for key, val := range a.MetricsGauge {
		valFloat64 := float64(val)

		msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
		heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)

		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
		chanMatrics <- metrica

		if cap(chanMatrics) != 0 && len(chanMatrics) == cap(chanMatrics) {
			for mt := range chanMatrics {
				allMetrics = append(allMetrics, mt)
				if len(chanMatrics) == 0 {
					break
				}
			}
			gziParrMterics, err := prepareMetrics(&allMetrics)
			if err != nil {
				constants.Logger.ErrorLog(err)
			}
			if err := a.Post2Server(&gziParrMterics); err != nil {
				constants.Logger.ErrorLog(err)
			}
			allMetrics = make(emtyArrMetrics, 0)
		}
	}

	cPollCount := repository.Counter(a.PollCount)
	msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.PollCount)
	heshVal := cryptohash.HeshSHA256(msg, a.Cfg.Key)
	chanMatrics <- encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.PollCount, Hash: heshVal}
	for mt := range chanMatrics {
		allMetrics = append(allMetrics, mt)
		if len(chanMatrics) == 0 {
			break
		}
	}

	gziparrMetrics, err := prepareMetrics(&allMetrics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	if err := a.Post2Server(&gziparrMetrics); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func main() {

	agent := agent{
		Cfg:          environment.SetConfigAgent(),
		MetricsGauge: make(MetricsGauge),
		PollCount:    0,
	}

	updateTicker := time.NewTicker(agent.Cfg.PollInterval)
	updateOtherTicker := time.NewTicker(agent.Cfg.PollInterval)
	reportTicker := time.NewTicker(agent.Cfg.ReportInterval)

	for {
		select {
		case <-updateTicker.C:
			agent.metrixScan()
		case <-updateOtherTicker.C:
			agent.metrixOtherScan()
		case <-reportTicker.C:
			agent.MakeRequest()
		}
	}

}

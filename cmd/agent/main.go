package main

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/Config"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricsGauge map[string]repository.Gauge

type Agent struct {
	MetricsGauge MetricsGauge
	PollCount    int64
	Cfg          Config.AgentConfig
}

func (a Agent) fillMetric(mem *runtime.MemStats) {

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

func (a Agent) metrixScan() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	a.fillMetric(&mem)
}

func (a Agent) Post2Server(arrMterica *[]byte) error {

	addressPost := fmt.Sprintf("http://%s/update", a.Cfg.Address)
	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(*arrMterica))
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("-------ошибка отправки данных на сервер (1)")
	}
	req.Header.Set("Content-Type", "application/json")
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("-------ошибка отправки данных на сервер (2)")
	}
	defer resp.Body.Close()

	return nil
}

func (a Agent) MakeRequest() {

	for key, val := range a.MetricsGauge {
		valFloat64 := float64(val)
		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64}
		arrMterica, err := metrica.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if err := a.Post2Server(&arrMterica); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	cPollCount := repository.Counter(a.PollCount)
	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.PollCount}
	arrMterica, err := metrica.MarshalMetrica()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := a.Post2Server(&arrMterica); err != nil {
		fmt.Println(err.Error())
		return
	}

}

func main() {

	agent := Agent{}
	agent.Cfg = Config.SetConfigAgent()
	agent.MetricsGauge = make(MetricsGauge)

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

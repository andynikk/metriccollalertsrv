package main

import (
	"bytes"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/caarlos0/env/v6"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

const (
//pollInterval   = 2
//reportInterval = 10
//msgFormat = "%s/update/%s/%s/%v"
)

type Config struct {
	address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	reportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
	rollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
}

var Cfg = Config{}

type MetricsGauge = map[string]repository.Gauge

//type Metrics = map[string]repository.Gauge

var PollCount int64

func fillMetric(metric MetricsGauge, mem *runtime.MemStats) {

	metric["Alloc"] = repository.Gauge(mem.Alloc)
	metric["BuckHashSys"] = repository.Gauge(mem.BuckHashSys)
	metric["GCCPUFraction"] = repository.Gauge(mem.GCCPUFraction)
	metric["GCSys"] = repository.Gauge(mem.GCSys)
	metric["HeapAlloc"] = repository.Gauge(mem.HeapAlloc)
	metric["HeapIdle"] = repository.Gauge(mem.HeapIdle)
	metric["HeapInuse"] = repository.Gauge(mem.HeapInuse)
	metric["HeapObjects"] = repository.Gauge(mem.HeapObjects)
	metric["HeapReleased"] = repository.Gauge(mem.HeapReleased)
	metric["HeapSys"] = repository.Gauge(mem.HeapSys)
	metric["LastGC"] = repository.Gauge(mem.LastGC)
	metric["Lookups"] = repository.Gauge(mem.Lookups)
	metric["MCacheInuse"] = repository.Gauge(mem.MCacheInuse)
	metric["MCacheSys"] = repository.Gauge(mem.MCacheSys)
	metric["MSpanInuse"] = repository.Gauge(mem.MSpanInuse)
	metric["MSpanSys"] = repository.Gauge(mem.MSpanSys)
	metric["Mallocs"] = repository.Gauge(mem.Mallocs)
	metric["NextGC"] = repository.Gauge(mem.NextGC)
	metric["NumForcedGC"] = repository.Gauge(mem.NumForcedGC)
	metric["NumGC"] = repository.Gauge(mem.NumGC)
	metric["OtherSys"] = repository.Gauge(mem.OtherSys)
	metric["PauseTotalNs"] = repository.Gauge(mem.PauseTotalNs)
	metric["StackInuse"] = repository.Gauge(mem.StackInuse)
	metric["StackSys"] = repository.Gauge(mem.StackSys)
	metric["Sys"] = repository.Gauge(mem.Sys)
	metric["TotalAlloc"] = repository.Gauge(mem.TotalAlloc)
	metric["RandomValue"] = repository.Gauge(rand.Float64())

	PollCount = PollCount + 1

}

func memThresholds(metric MetricsGauge) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fillMetric(metric, &mem)

}

func metrixScan(metric MetricsGauge) {

	memThresholds(metric)
}

func MakeRequest(metric MetricsGauge) {

	//message := makeMsg(metric)
	//rn := strings.NewReader(message)

	msg := "http://" + Cfg.address + "/update"
	//msg := "http://localhost:8080/update"
	for key, val := range metric {
		//msg := fmt.Sprintf(msgFormat, consts.AddressServer, val.Type(), key, val)
		valFloat64 := val.Float64()
		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64}
		arrMterica, err := metrica.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		req, err := http.NewRequest("POST", msg, bytes.NewBuffer(arrMterica))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			fmt.Println(err.Error())
		}
		defer req.Body.Close()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer resp.Body.Close()
	}

	cPollCount := repository.Counter(PollCount)
	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &PollCount}
	arrMterica, err := metrica.MarshalMetrica()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	req, err := http.NewRequest("POST", msg, bytes.NewBuffer(arrMterica))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
}

func main() {

	err := env.Parse(&Cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	metric := make(MetricsGauge)

	updateTicker := time.NewTicker(time.Duration(Cfg.rollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(Cfg.reportInterval) * time.Second)

	for {
		select {
		case <-updateTicker.C:
			metrixScan(metric)
		case <-reportTicker.C:
			MakeRequest(metric)
		}
	}

}

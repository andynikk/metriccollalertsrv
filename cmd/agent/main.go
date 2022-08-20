package main

import (
	"bytes"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/consts"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

const (
	pollInterval   = 2
	reportInterval = 10
	msgFormat      = "%s/update/%s/%s/%v"
)

//type Metrics = map[string]repository.Gauge

var PollCount int64

func fillMetric(metric encoding.MetricsGauge, mem *runtime.MemStats) {

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

func memThresholds(metric encoding.MetricsGauge) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fillMetric(metric, &mem)

}

func metrixScan(metric encoding.MetricsGauge) {

	memThresholds(metric)
}

func MakeRequest(metric encoding.MetricsGauge) {

	//message := makeMsg(metric)
	//rn := strings.NewReader(message)

	msg := consts.AddressServer + "/update"
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
	msg1 := fmt.Sprintf(msgFormat, consts.AddressServer, cPollCount.Type(), "PollCount", cPollCount)
	rn := strings.NewReader(msg1)

	resp, err := http.Post(msg1, "text/plain", rn)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

}

func main() {

	metric := make(encoding.MetricsGauge)

	updateTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)

	for {
		select {
		case <-updateTicker.C:
			metrixScan(metric)
		case <-reportTicker.C:
			MakeRequest(metric)
		}
	}

}

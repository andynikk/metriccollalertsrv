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

type Config struct {
	ADDRESS         string `env:"ADDRESS" envDefault:"localhost:8080"`
	REPORT_INTERVAL int64  `env:"reportInterval" envDefault:"10"`
	POLL_INTERVAL   int64  `env:"rollInterval" envDefault:"2"`
}

var Cfg = Config{}

type MetricsGauge = map[string]repository.Gauge

//type Metrics = map[string]repository.Gauge

var PollCount int64

func fillMetric(metric MetricsGauge, mem *runtime.MemStats) {

	metric["Alloc"] = repository.Gauge(mem.Alloc)
	metric["BuckHashSys"] = repository.Gauge(mem.BuckHashSys)
	metric["Frees"] = repository.Gauge(mem.Frees)
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

	msg := "http://" + Cfg.ADDRESS + "/update"

	//fmt.Printf("Сервер:", msg, "\n")
	//fmt.Printf("Количество метрик: %d\n", len(metric))

	for key, val := range metric {
		valFloat64 := val.Float64()
		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64}
		arrMterica, err := metrica.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		//fmt.Println("Отправка метрики", metrica.ID, metrica.MType, metrica.Value, metrica.Delta)
		resp, err := http.Post(msg, "application/json", bytes.NewReader(arrMterica))
		if err != nil {
			//fmt.Println("Ошибка отправки JSON метрики", string(arrMterica))
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

	//fmt.Println("Отправка метрики", metrica.ID, metrica.MType, metrica.Value, metrica.Delta)
	resp, err := http.Post(msg, "application/json", bytes.NewReader(arrMterica))
	if err != nil {
		//fmt.Println("Ошибка отправки метрики", metrica.ID, metrica.MType, metrica.Value, metrica.Delta)
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

	updateTicker := time.NewTicker(time.Duration(Cfg.POLL_INTERVAL) * time.Second)
	reportTicker := time.NewTicker(time.Duration(Cfg.REPORT_INTERVAL) * time.Second)

	for {
		select {
		case <-updateTicker.C:
			metrixScan(metric)
		case <-reportTicker.C:
			MakeRequest(metric)
		}
	}

}

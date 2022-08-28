package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type ConfigENV struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type Config struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type MetricsGauge = map[string]repository.Gauge

var PollCount int64
var Cfg = Config{}

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

func CompressAndPost(arrMterica *[]byte) error {

	var bytMterica []byte
	b := bytes.NewBuffer(*arrMterica).Bytes()
	bytMterica = append(bytMterica, b...)
	compData, err := compression.Compress(bytMterica)
	if err != nil {
		fmt.Println(compData)
		return errors.New("ошибка архивации данных")
	}

	compDataBytes := bytes.NewReader(compData)
	req, err := http.NewRequest("POST", "http://"+Cfg.Address+"/update", compDataBytes)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New("ошибка отправки данных на сервер (POST)")
	}
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Encoding", "gzip")
	//req.Header.Set("Accept-Encoding", "gzip")

	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error(), string(compData))
		return errors.New("ошибка отправки данных на сервер (Do)")
	}
	defer resp.Body.Close()

	return nil
}

func MakeRequest(metric MetricsGauge) {

	msg := "http://" + Cfg.Address + "/update"

	for key, val := range metric {
		valFloat64 := val.Float64()
		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64}
		arrMterica, err := metrica.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		//fmt.Println("-------------", metrica, arrMterica)
		//if err := CompressAndPost(&arrMterica); err != nil {
		//	fmt.Println(err.Error())
		//	continue
		//}
		resp, err := http.Post(msg, "application/json", bytes.NewReader(arrMterica))
		if err != nil {
			fmt.Println(err.Error())
		}
		defer resp.Body.Close()

		//var bytMterica []byte
		//b := bytes.NewBuffer(arrMterica).Bytes()
		//bytMterica = append(bytMterica, b...)
		//compData, err := compression.Compress(bytMterica)
		//if err != nil {
		//	fmt.Println(compData)
		//	continue
		//}

		req, err := http.NewRequest("POST", msg, bytes.NewBuffer(arrMterica))
		if err != nil {
			fmt.Println(err.Error())
		}
		req.Header.Set("Content-Type", "application/json")

		//req.Header.Set("Content-Encoding", "gzip")

		//if err != nil {
		//	fmt.Println(err.Error())
		//}
		//defer req.Body.Close()
		//
		//client := &http.Client{}
		//resp, err := client.Do(req)
		//if err != nil {
		//	fmt.Println(err.Error())
		//}
		//defer resp.Body.Close()
		//
		//defer resp.Body.Close()
		//resp.Body.Close()
	}

	cPollCount := repository.Counter(PollCount)
	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &PollCount}
	arrMterica, err := metrica.MarshalMetrica()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//if err := CompressAndPost(&arrMterica); err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}

	resp, err := http.Post(msg, "application/json", bytes.NewReader(arrMterica))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	//resp.Body.Close()
	//
	//var bytMterica []byte
	//b := bytes.NewBuffer(arrMterica).Bytes()
	//bytMterica = append(bytMterica, b...)
	//compData, err := compression.Compress(bytMterica)
	//if err != nil {
	//	fmt.Println(compData)
	//	return
	//}
	//
	////req, err := http.NewRequest("POST", msg, bytes.NewBuffer(arrMterica))
	//req, err := http.NewRequest("POST", msg, bytes.NewReader(compData))
	//
	//req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Encoding", "gzip")
	//
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//defer req.Body.Close()
	//
	//client := &http.Client{}
	//resp, err := client.Do(req)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//defer resp.Body.Close()

}

func main() {

	addressPtr := flag.String("a", "localhost:8080", "имя сервера")
	reportIntervalPtr := flag.Duration("r", 10*time.Second, "интервал отправки на сервер")
	pollIntervalPtr := flag.Duration("p", 2*time.Second, "интервал сбора метрик")
	flag.Parse()

	var cfgENV ConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	addressServ := ""
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		addressServ = cfgENV.Address
	} else {
		addressServ = *addressPtr
	}

	var reportIntervalMetric time.Duration
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportIntervalMetric = cfgENV.ReportInterval
	} else {
		reportIntervalMetric = *reportIntervalPtr
	}

	var pollIntervalMetrics time.Duration
	if _, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollIntervalMetrics = cfgENV.PollInterval
	} else {
		pollIntervalMetrics = *pollIntervalPtr
	}

	Cfg = Config{
		Address:        addressServ,
		ReportInterval: reportIntervalMetric,
		PollInterval:   pollIntervalMetrics,
	}

	metric := make(MetricsGauge)

	updateTicker := time.NewTicker(Cfg.PollInterval)   // * time.Second)
	reportTicker := time.NewTicker(Cfg.ReportInterval) // * time.Second)

	for {
		select {
		case <-updateTicker.C:
			metrixScan(metric)
		case <-reportTicker.C:
			MakeRequest(metric)
		}
	}

}

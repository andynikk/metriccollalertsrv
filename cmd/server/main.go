package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

var servMetrixStats MemStats

func (g gauge) Type() string {
	return fmt.Sprintf("%T", g)
}

func (c counter) Type() string {
	return fmt.Sprintf("%T", c)
}

type MemStats struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	RandomValue   gauge
	PollCount     counter
}

func valueGauge(val string) float64 {
	arrGauge := strings.Split(val, "/")
	valGauge := arrGauge[len(arrGauge)-1]
	fltGauge, err := strconv.ParseFloat(valGauge, 64)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return fltGauge
}

func valueCounter(val string) int64 {

	arrCounter := strings.Split(val, "/")
	valCounter := arrCounter[len(arrCounter)-1]
	intCounter, err := strconv.ParseInt(valCounter, 10, 64)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return intCounter
}

func changeMetrx(messageRaz []string, memStats *MemStats) {

	//var msg []string

	for _, val := range messageRaz {

		if strings.Contains(val, servMetrixStats.Alloc.Type()) {
			valueGauge := valueGauge(val)
			if strings.Contains(val, "Alloc") {
				servMetrixStats.Alloc = gauge(valueGauge)
			} else if strings.Contains(val, "BuckHashSys") {
				servMetrixStats.BuckHashSys = gauge(valueGauge)
			} else if strings.Contains(val, "Frees") {
				servMetrixStats.Frees = gauge(valueGauge)
			} else if strings.Contains(val, "GCCPUFraction") {
				servMetrixStats.GCCPUFraction = gauge(valueGauge)
			} else if strings.Contains(val, "GCSys") {
				servMetrixStats.GCSys = gauge(valueGauge)
			} else if strings.Contains(val, "HeapAlloc") {
				servMetrixStats.HeapAlloc = gauge(valueGauge)
			} else if strings.Contains(val, "HeapIdle") {
				servMetrixStats.HeapIdle = gauge(valueGauge)
			} else if strings.Contains(val, "HeapInuse") {
				servMetrixStats.HeapInuse = gauge(valueGauge)
			} else if strings.Contains(val, "HeapObjects") {
				servMetrixStats.HeapObjects = gauge(valueGauge)
			} else if strings.Contains(val, "HeapReleased") {
				servMetrixStats.HeapReleased = gauge(valueGauge)
			} else if strings.Contains(val, "HeapSys") {
				servMetrixStats.HeapSys = gauge(valueGauge)
			} else if strings.Contains(val, "LastGC") {
				servMetrixStats.LastGC = gauge(valueGauge)
			} else if strings.Contains(val, "Lookups") {
				servMetrixStats.Lookups = gauge(valueGauge)
			} else if strings.Contains(val, "MCacheInuse") {
				servMetrixStats.MCacheInuse = gauge(valueGauge)
			} else if strings.Contains(val, "MCacheSys") {
				servMetrixStats.MCacheSys = gauge(valueGauge)
			} else if strings.Contains(val, "MSpanInuse") {
				servMetrixStats.MSpanInuse = gauge(valueGauge)
			} else if strings.Contains(val, "MSpanSys") {
				servMetrixStats.MSpanSys = gauge(valueGauge)
			} else if strings.Contains(val, "Mallocs") {
				servMetrixStats.Mallocs = gauge(valueGauge)
			} else if strings.Contains(val, "NextGC") {
				servMetrixStats.NextGC = gauge(valueGauge)
			} else if strings.Contains(val, "NumForcedGC") {
				servMetrixStats.NumForcedGC = gauge(valueGauge)
			} else if strings.Contains(val, "NumGC") {
				servMetrixStats.NumGC = gauge(valueGauge)
			} else if strings.Contains(val, "OtherSys") {
				servMetrixStats.OtherSys = gauge(valueGauge)
			} else if strings.Contains(val, "PauseTotalNs") {
				servMetrixStats.PauseTotalNs = gauge(valueGauge)
			} else if strings.Contains(val, "StackInuse") {
				servMetrixStats.StackInuse = gauge(valueGauge)
			} else if strings.Contains(val, "StackSys") {
				servMetrixStats.StackSys = gauge(valueGauge)
			} else if strings.Contains(val, "Sys") {
				servMetrixStats.Sys = gauge(valueGauge)
			} else if strings.Contains(val, "TotalAlloc") {
				servMetrixStats.TotalAlloc = gauge(valueGauge)
			} else if strings.Contains(val, "RandomValue") {
				servMetrixStats.RandomValue = gauge(valueGauge)
			}
		} else if strings.Contains(val, memStats.PollCount.Type()) {
			valueCounter := valueCounter(val)
			if strings.Contains(val, "PollCount") {

				servMetrixStats.PollCount = servMetrixStats.PollCount + counter(valueCounter)

			}
		}
	}

}

func handleFunc(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		fmt.Println("Not POST")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "err %q\n", err)
		return
	}

	message := string(b)
	messageRaz := strings.Split(message, "\n")

	changeMetrx(messageRaz, &servMetrixStats)

}

func main() {

	http.HandleFunc("/", handleFunc)
	http.ListenAndServe(":8080", nil)

	servMetrixStats.PollCount = 0

}

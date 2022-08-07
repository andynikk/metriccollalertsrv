package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

type gauge float64
type counter int64

func (g gauge) Type() string {
	return "gauge"
}

func (c counter) Type() string {
	return "counter"
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

func fillGauge(memStats *MemStats, mem *runtime.MemStats) {
	memStats.Alloc = gauge(mem.Alloc)
	memStats.BuckHashSys = gauge(mem.BuckHashSys)
	memStats.Frees = gauge(mem.Frees)
	memStats.GCCPUFraction = gauge(mem.GCCPUFraction)
	memStats.GCSys = gauge(mem.GCSys)
	memStats.HeapAlloc = gauge(mem.HeapAlloc)
	memStats.HeapIdle = gauge(mem.HeapIdle)
	memStats.HeapInuse = gauge(mem.HeapInuse)
	memStats.HeapObjects = gauge(mem.HeapObjects)
	memStats.HeapReleased = gauge(mem.HeapReleased)
	memStats.HeapSys = gauge(mem.HeapSys)
	memStats.LastGC = gauge(mem.LastGC)
	memStats.Lookups = gauge(mem.Lookups)
	memStats.MCacheInuse = gauge(mem.MCacheInuse)
	memStats.MCacheSys = gauge(mem.MCacheSys)
	memStats.MSpanInuse = gauge(mem.MSpanInuse)
	memStats.MSpanSys = gauge(mem.MSpanSys)
	memStats.Mallocs = gauge(mem.Mallocs)
	memStats.NextGC = gauge(mem.NextGC)
	memStats.NumForcedGC = gauge(mem.NumForcedGC)
	memStats.NumGC = gauge(mem.NumGC)
	memStats.OtherSys = gauge(mem.OtherSys)
	memStats.PauseTotalNs = gauge(mem.PauseTotalNs)
	memStats.StackInuse = gauge(mem.StackInuse)
	memStats.StackSys = gauge(mem.StackSys)
	memStats.Sys = gauge(mem.Sys)
	memStats.TotalAlloc = gauge(mem.TotalAlloc)
	memStats.RandomValue = gauge(rand.Float64())
}

func fillCounter(memStats *MemStats) {
	memStats.PollCount = counter(memStats.PollCount + 1)
}

func memThresholds(memStats *MemStats) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	//fmt.Println(memStats)

	fillGauge(memStats, &mem)
	fillCounter(memStats)

}

func metrixScan(memStats *MemStats) {

	memThresholds(memStats)
}

func makeMsg(memStats MemStats) string {
	const adresServer = "localhost:8080"
	const msgFormat = "http://%s/update/%s/%s/%v"
	//const msgFormat = "update/%s/%s/%v"

	var msg []string

	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Alloc.Type(), "Alloc", memStats.Alloc))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.BuckHashSys.Type(), "BuckHashSys", memStats.BuckHashSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Frees.Type(), "Frees", memStats.Frees))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.GCCPUFraction.Type(), "GCCPUFraction", memStats.GCCPUFraction))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.GCSys.Type(), "GCSys", memStats.GCSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapAlloc.Type(), "HeapAlloc", memStats.HeapAlloc))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapIdle.Type(), "HeapIdle", memStats.HeapIdle))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapInuse.Type(), "HeapInuse", memStats.HeapInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapObjects.Type(), "HeapObjects", memStats.HeapObjects))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapReleased.Type(), "HeapReleased", memStats.HeapReleased))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.HeapSys.Type(), "HeapSys", memStats.HeapSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.LastGC.Type(), "LastGC", memStats.LastGC))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Lookups.Type(), "Lookups", memStats.Lookups))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.MCacheInuse.Type(), "MCacheInuse", memStats.MCacheInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.MCacheSys.Type(), "MCacheSys", memStats.MCacheSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.MSpanInuse.Type(), "MSpanInuse", memStats.MSpanInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.MSpanSys.Type(), "MSpanSys", memStats.MSpanSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Mallocs.Type(), "Mallocs", memStats.Mallocs))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.NextGC.Type(), "NextGC", memStats.NextGC))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.NumForcedGC.Type(), "NumForcedGC", memStats.NumForcedGC))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.NumGC.Type(), "NumGC", memStats.NumGC))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.OtherSys.Type(), "OtherSys", memStats.OtherSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.PauseTotalNs.Type(), "PauseTotalNs", memStats.PauseTotalNs))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.StackInuse.Type(), "StackInuse", memStats.StackInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.StackSys.Type(), "StackSys", memStats.StackSys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Sys.Type(), "Sys", memStats.Sys))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.TotalAlloc.Type(), "TotalAlloc", memStats.TotalAlloc))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.RandomValue.Type(), "RandomValue", memStats.RandomValue))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.PollCount.Type(), "PollCount", memStats.PollCount))

	return strings.Join(msg, "\n")
}

func MakeRequest(memStats *MemStats) {

	message := makeMsg(*memStats)
	r := strings.NewReader(message)

	//r := strings.NewReader("update/gauge/testGauge/100")

	resp, err := http.Post("http://localhost:8080", "text/plain", r)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	//fmt.Println(resp.Status)
	//fmt.Println("Сообщение: \n" + message + "\nотправлено успешно")

}

func startMetric(memStats *MemStats) {
	ticker := time.NewTicker(pollInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				break
			}
			metrixScan(memStats)
		}
	}
}

func startSender(memStats *MemStats) {
	ticker := time.NewTicker(reportInterval * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			MakeRequest(memStats)
		}
	}
}

func main() {

	memStats := MemStats{}
	go startMetric(&memStats)
	go startSender(&memStats)

	exit := make(chan os.Signal, 1024)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit

	//start := time.Now()
	//
	//ticker := time.NewTicker(pollInterval * time.Second)
	//defer ticker.Stop()
	//
	//done := make(chan bool)
	//go func() {
	//	time.Sleep(reportInterval * time.Second)
	//	done <- true
	//	//fmt.Println(1)
	//}()
	//for {
	//	select {
	//	case <-done:
	//		//fmt.Println(memStats)
	//
	//		MakeRequest(&memStats)
	//
	//		go func() {
	//			time.Sleep(reportInterval * time.Second)
	//			done <- true
	//			//fmt.Println(1)
	//		}()
	//		//return
	//	case t := <-ticker.C:
	//		metrixScan(&memStats)
	//		t.Sub(start).Seconds()
	//		//fmt.Println(2)
	//	}
	//}
}

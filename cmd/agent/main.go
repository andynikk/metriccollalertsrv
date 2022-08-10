package main

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	addressServer  = "http://localhost:8080"
	msgFormat      = "%s/update/%s/%s/%v"
)

type MemStats struct {
	Alloc         repository.Gauge
	BuckHashSys   repository.Gauge
	Frees         repository.Gauge
	GCCPUFraction repository.Gauge
	GCSys         repository.Gauge
	HeapAlloc     repository.Gauge
	HeapIdle      repository.Gauge
	HeapInuse     repository.Gauge
	HeapObjects   repository.Gauge
	HeapReleased  repository.Gauge
	HeapSys       repository.Gauge
	LastGC        repository.Gauge
	Lookups       repository.Gauge
	MCacheInuse   repository.Gauge
	MCacheSys     repository.Gauge
	MSpanInuse    repository.Gauge
	MSpanSys      repository.Gauge
	Mallocs       repository.Gauge
	NextGC        repository.Gauge
	NumForcedGC   repository.Gauge
	NumGC         repository.Gauge
	OtherSys      repository.Gauge
	PauseTotalNs  repository.Gauge
	StackInuse    repository.Gauge
	StackSys      repository.Gauge
	Sys           repository.Gauge
	TotalAlloc    repository.Gauge
	RandomValue   repository.Gauge
	PollCount     repository.Counter
}

func fillGauge(memStats *MemStats, mem *runtime.MemStats) {
	memStats.Alloc = repository.Gauge(mem.Alloc)
	memStats.BuckHashSys = repository.Gauge(mem.BuckHashSys)
	memStats.Frees = repository.Gauge(mem.Frees)
	memStats.GCCPUFraction = repository.Gauge(mem.GCCPUFraction)
	memStats.GCSys = repository.Gauge(mem.GCSys)
	memStats.HeapAlloc = repository.Gauge(mem.HeapAlloc)
	memStats.HeapIdle = repository.Gauge(mem.HeapIdle)
	memStats.HeapInuse = repository.Gauge(mem.HeapInuse)
	memStats.HeapObjects = repository.Gauge(mem.HeapObjects)
	memStats.HeapReleased = repository.Gauge(mem.HeapReleased)
	memStats.HeapSys = repository.Gauge(mem.HeapSys)
	memStats.LastGC = repository.Gauge(mem.LastGC)
	memStats.Lookups = repository.Gauge(mem.Lookups)
	memStats.MCacheInuse = repository.Gauge(mem.MCacheInuse)
	memStats.MCacheSys = repository.Gauge(mem.MCacheSys)
	memStats.MSpanInuse = repository.Gauge(mem.MSpanInuse)
	memStats.MSpanSys = repository.Gauge(mem.MSpanSys)
	memStats.Mallocs = repository.Gauge(mem.Mallocs)
	memStats.NextGC = repository.Gauge(mem.NextGC)
	memStats.NumForcedGC = repository.Gauge(mem.NumForcedGC)
	memStats.NumGC = repository.Gauge(mem.NumGC)
	memStats.OtherSys = repository.Gauge(mem.OtherSys)
	memStats.PauseTotalNs = repository.Gauge(mem.PauseTotalNs)
	memStats.StackInuse = repository.Gauge(mem.StackInuse)
	memStats.StackSys = repository.Gauge(mem.StackSys)
	memStats.Sys = repository.Gauge(mem.Sys)
	memStats.TotalAlloc = repository.Gauge(mem.TotalAlloc)
	memStats.RandomValue = repository.Gauge(rand.Float64())
}

func fillCounter(memStats *MemStats) {

	memStats.PollCount = repository.Counter(memStats.PollCount + 1)
}

func memThresholds(memStats *MemStats) {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fillGauge(memStats, &mem)
	fillCounter(memStats)

}

func metrixScan(memStats *MemStats) {

	memThresholds(memStats)
}

func makeMsg(memStats *MemStats) string {

	var msg []string

	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.Alloc.Type(), "Alloc", memStats.Alloc))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.BuckHashSys.Type(), "BuckHashSys", memStats.BuckHashSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.Frees.Type(), "Frees", memStats.Frees))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.GCCPUFraction.Type(), "GCCPUFraction", memStats.GCCPUFraction))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.GCSys.Type(), "GCSys", memStats.GCSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapAlloc.Type(), "HeapAlloc", memStats.HeapAlloc))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapIdle.Type(), "HeapIdle", memStats.HeapIdle))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapInuse.Type(), "HeapInuse", memStats.HeapInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapObjects.Type(), "HeapObjects", memStats.HeapObjects))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapReleased.Type(), "HeapReleased", memStats.HeapReleased))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.HeapSys.Type(), "HeapSys", memStats.HeapSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.LastGC.Type(), "LastGC", memStats.LastGC))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.Lookups.Type(), "Lookups", memStats.Lookups))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.MCacheInuse.Type(), "MCacheInuse", memStats.MCacheInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.MCacheSys.Type(), "MCacheSys", memStats.MCacheSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.MSpanInuse.Type(), "MSpanInuse", memStats.MSpanInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.MSpanSys.Type(), "MSpanSys", memStats.MSpanSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.Mallocs.Type(), "Mallocs", memStats.Mallocs))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.NextGC.Type(), "NextGC", memStats.NextGC))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.NumForcedGC.Type(), "NumForcedGC", memStats.NumForcedGC))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.NumGC.Type(), "NumGC", memStats.NumGC))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.OtherSys.Type(), "OtherSys", memStats.OtherSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.PauseTotalNs.Type(), "PauseTotalNs", memStats.PauseTotalNs))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.StackInuse.Type(), "StackInuse", memStats.StackInuse))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.StackSys.Type(), "StackSys", memStats.StackSys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.Sys.Type(), "Sys", memStats.Sys))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.TotalAlloc.Type(), "TotalAlloc", memStats.TotalAlloc))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.RandomValue.Type(), "RandomValue", memStats.RandomValue))
	msg = append(msg, fmt.Sprintf(msgFormat, addressServer, memStats.PollCount.Type(), "PollCount", memStats.PollCount))

	return strings.Join(msg, "\n")
}

func MakeRequest(memStats *MemStats) {

	message := makeMsg(memStats)
	rn := strings.NewReader(message)

	//r := strings.NewReader("update/gauge/Alloc/100")

	resp, err := http.Post(addressServer, "text/plain", rn)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	//fmt.Println(resp.Status)
	//fmt.Println("Сообщение: \n" + message + "\nотправлено успешно")

}

func main() {

	memStats := MemStats{}

	updateTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)

	for {
		select {
		case <-updateTicker.C:
			metrixScan(&memStats)
		case <-reportTicker.C:
			MakeRequest(&memStats)
		}
	}

	//go startMetric(&memStats)
	//go startSender(&memStats)
	//
	//exit := make(chan os.Signal)
	//signal.Notify(exit, os.Interrupt, os.Kill)
	//<-exit

}

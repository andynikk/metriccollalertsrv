package main

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/models"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	adresServer    = "localhost:8080"
)

type MemStats struct {
	Alloc         models.Gauge
	BuckHashSys   models.Gauge
	Frees         models.Gauge
	GCCPUFraction models.Gauge
	GCSys         models.Gauge
	HeapAlloc     models.Gauge
	HeapIdle      models.Gauge
	HeapInuse     models.Gauge
	HeapObjects   models.Gauge
	HeapReleased  models.Gauge
	HeapSys       models.Gauge
	LastGC        models.Gauge
	Lookups       models.Gauge
	MCacheInuse   models.Gauge
	MCacheSys     models.Gauge
	MSpanInuse    models.Gauge
	MSpanSys      models.Gauge
	Mallocs       models.Gauge
	NextGC        models.Gauge
	NumForcedGC   models.Gauge
	NumGC         models.Gauge
	OtherSys      models.Gauge
	PauseTotalNs  models.Gauge
	StackInuse    models.Gauge
	StackSys      models.Gauge
	Sys           models.Gauge
	TotalAlloc    models.Gauge
	RandomValue   models.Gauge
	PollCount     models.Counter
}

func fillGauge(memStats *MemStats, mem *runtime.MemStats) {
	memStats.Alloc = models.Gauge(mem.Alloc)
	memStats.BuckHashSys = models.Gauge(mem.BuckHashSys)
	memStats.Frees = models.Gauge(mem.Frees)
	memStats.GCCPUFraction = models.Gauge(mem.GCCPUFraction)
	memStats.GCSys = models.Gauge(mem.GCSys)
	memStats.HeapAlloc = models.Gauge(mem.HeapAlloc)
	memStats.HeapIdle = models.Gauge(mem.HeapIdle)
	memStats.HeapInuse = models.Gauge(mem.HeapInuse)
	memStats.HeapObjects = models.Gauge(mem.HeapObjects)
	memStats.HeapReleased = models.Gauge(mem.HeapReleased)
	memStats.HeapSys = models.Gauge(mem.HeapSys)
	memStats.LastGC = models.Gauge(mem.LastGC)
	memStats.Lookups = models.Gauge(mem.Lookups)
	memStats.MCacheInuse = models.Gauge(mem.MCacheInuse)
	memStats.MCacheSys = models.Gauge(mem.MCacheSys)
	memStats.MSpanInuse = models.Gauge(mem.MSpanInuse)
	memStats.MSpanSys = models.Gauge(mem.MSpanSys)
	memStats.Mallocs = models.Gauge(mem.Mallocs)
	memStats.NextGC = models.Gauge(mem.NextGC)
	memStats.NumForcedGC = models.Gauge(mem.NumForcedGC)
	memStats.NumGC = models.Gauge(mem.NumGC)
	memStats.OtherSys = models.Gauge(mem.OtherSys)
	memStats.PauseTotalNs = models.Gauge(mem.PauseTotalNs)
	memStats.StackInuse = models.Gauge(mem.StackInuse)
	memStats.StackSys = models.Gauge(mem.StackSys)
	memStats.Sys = models.Gauge(mem.Sys)
	memStats.TotalAlloc = models.Gauge(mem.TotalAlloc)
	memStats.RandomValue = models.Gauge(rand.Float64())
}

func fillCounter(memStats *MemStats) {

	memStats.PollCount = models.Counter(memStats.PollCount + 1)
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

func MakeRequest(memStats MemStats) {

	message := makeMsg(memStats)
	r := strings.NewReader(message)

	//r := strings.NewReader("update/models.Gauge/testmodels.Gauge/100")

	resp, err := http.Post(adresServer, "text/plain", r)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	//fmt.Println(resp.Status)
	//fmt.Println("Сообщение: \n" + message + "\nотправлено успешно")

}

func main() {

	memStats := MemStats{}

	updateTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)

	select {
	case <-updateTicker.C:
		metrixScan(&memStats)
	case <-reportTicker.C:
		MakeRequest(memStats)
	}

}

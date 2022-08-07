package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"io/ioutil"
	"log"
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

func (c counter) String() string {
	return fmt.Sprintf("%d", int64(c))
}

func (g gauge) String() string {
	fg := float64(g)
	return fmt.Sprintf("%g", fg)
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

func valStrMetrics(strMetrix string, numElArr int64) string {
	arrMetrics := strings.Split(strMetrix, "/")
	valMetrics := arrMetrics[numElArr]

	return valMetrics
}

func changeMetrx(messageRaz []string, memStats *MemStats) {

	//var msg []string

	for _, val := range messageRaz {
		typeMetric := valStrMetrics(val, 4)
		nameMetric := valStrMetrics(val, 5)

		if typeMetric == "gauge" {
			valueGauge := valueGauge(val)
			switch nameMetric {
			case "Alloc":
				servMetrixStats.Alloc = gauge(valueGauge)
			case "BuckHashSys":
				servMetrixStats.BuckHashSys = gauge(valueGauge)

			case "Frees":
				servMetrixStats.Frees = gauge(valueGauge)

			case "GCCPUFraction":
				servMetrixStats.GCCPUFraction = gauge(valueGauge)

			case "GCSys":
				servMetrixStats.GCSys = gauge(valueGauge)

			case "HeapAlloc":
				servMetrixStats.HeapAlloc = gauge(valueGauge)

			case "HeapIdle":
				servMetrixStats.HeapIdle = gauge(valueGauge)

			case "HeapInuse":
				servMetrixStats.HeapInuse = gauge(valueGauge)

			case "HeapObjects":
				servMetrixStats.HeapObjects = gauge(valueGauge)

			case "HeapReleased":
				servMetrixStats.HeapReleased = gauge(valueGauge)

			case "HeapSys":
				servMetrixStats.HeapSys = gauge(valueGauge)

			case "LastGC":
				servMetrixStats.LastGC = gauge(valueGauge)

			case "Lookups":
				servMetrixStats.Lookups = gauge(valueGauge)

			case "MCacheInuse":
				servMetrixStats.MCacheInuse = gauge(valueGauge)

			case "MCacheSys":
				servMetrixStats.MCacheSys = gauge(valueGauge)

			case "MSpanInuse":
				servMetrixStats.MSpanInuse = gauge(valueGauge)

			case "MSpanSys":
				servMetrixStats.MSpanSys = gauge(valueGauge)

			case "Mallocs":
				servMetrixStats.Mallocs = gauge(valueGauge)

			case "NextGC":
				servMetrixStats.NextGC = gauge(valueGauge)

			case "NumForcedGC":
				servMetrixStats.NumForcedGC = gauge(valueGauge)

			case "NumGC":
				servMetrixStats.NumGC = gauge(valueGauge)

			case "OtherSys":
				servMetrixStats.OtherSys = gauge(valueGauge)

			case "PauseTotalNs":
				servMetrixStats.PauseTotalNs = gauge(valueGauge)

			case "StackInuse":
				servMetrixStats.StackInuse = gauge(valueGauge)

			case "StackSys":
				servMetrixStats.StackSys = gauge(valueGauge)

			case "Sys":
				servMetrixStats.Sys = gauge(valueGauge)

			case "TotalAlloc":
				servMetrixStats.TotalAlloc = gauge(valueGauge)

			case "RandomValue":
				servMetrixStats.RandomValue = gauge(valueGauge)

			}
		} else if typeMetric == "counter" {
			valueCounter := valueCounter(val)
			switch nameMetric {
			case "PollCount":
				servMetrixStats.PollCount = servMetrixStats.PollCount + counter(valueCounter)
			}

		}
	}

}

func handleFunc(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		//fmt.Println("Not POST")
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

func notFound(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	if r.Method != "GET" {
		return
	}

	_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	if err != nil {
		panic(err)
	}
}

func textMetricsAndValue() string {
	const msgFormat = "%s = %s"

	var msg []string
	msg = append(msg, fmt.Sprintf(msgFormat, "Alloc", servMetrixStats.Alloc.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "BuckHashSys", servMetrixStats.BuckHashSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "Frees", servMetrixStats.Frees.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "GCCPUFraction", servMetrixStats.GCCPUFraction))
	msg = append(msg, fmt.Sprintf(msgFormat, "GCSys", servMetrixStats.GCSys))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapAlloc", servMetrixStats.HeapAlloc.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapIdle", servMetrixStats.HeapIdle.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapInuse", servMetrixStats.HeapInuse.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapObjects", servMetrixStats.HeapObjects.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapReleased", servMetrixStats.HeapReleased.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "HeapSys", servMetrixStats.HeapSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "LastGC", servMetrixStats.LastGC.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "Lookups", servMetrixStats.Lookups.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "MCacheInuse", servMetrixStats.MCacheInuse.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "MCacheSys", servMetrixStats.MCacheSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "MSpanInuse", servMetrixStats.MSpanInuse.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "MSpanSys", servMetrixStats.MSpanSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "Mallocs", servMetrixStats.Mallocs.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "NextGC", servMetrixStats.NextGC.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "NumForcedGC", servMetrixStats.NumForcedGC.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "NumGC", servMetrixStats.NumGC.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "OtherSys", servMetrixStats.OtherSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "PauseTotalNs", servMetrixStats.PauseTotalNs.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "StackInuse", servMetrixStats.StackInuse.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "StackSys", servMetrixStats.StackSys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "Sys", servMetrixStats.Sys.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "TotalAlloc", servMetrixStats.TotalAlloc.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "RandomValue", servMetrixStats.RandomValue.String()))
	msg = append(msg, fmt.Sprintf(msgFormat, "PollCount", servMetrixStats.PollCount.String()))

	return strings.Join(msg, "\n")
}

func getValueMetrics(typeVal string, val string) string {

	if typeVal != "gauge" && typeVal != "counter" {
		return "-"
	}

	switch val {

	case "Alloc":
		return servMetrixStats.Alloc.String()

	case "BuckHashSys":
		return servMetrixStats.BuckHashSys.String()

	case "Frees":
		return servMetrixStats.Frees.String()

	case "GCCPUFraction":
		return servMetrixStats.GCCPUFraction.String()

	case "GCSys":
		return servMetrixStats.GCSys.String()

	case "HeapAlloc":
		return servMetrixStats.HeapAlloc.String()

	case "HeapIdle":
		return servMetrixStats.HeapIdle.String()
	case "HeapInuse":
		return servMetrixStats.HeapInuse.String()

	case "HeapObjects":
		return servMetrixStats.HeapObjects.String()

	case "HeapReleased":
		return servMetrixStats.HeapReleased.String()
	case "HeapSys":
		return servMetrixStats.HeapSys.String()

	case "LastGC":
		return servMetrixStats.LastGC.String()

	case "Lookups":
		return servMetrixStats.Lookups.String()

	case "MCacheInuse":
		return servMetrixStats.MCacheInuse.String()

	case "MCacheSys":
		return servMetrixStats.MCacheSys.String()

	case "MSpanInuse":
		return servMetrixStats.MSpanInuse.String()

	case "MSpanSys":
		return servMetrixStats.MSpanSys.String()

	case "Mallocs":
		return servMetrixStats.Mallocs.String()

	case "NextGC":
		return servMetrixStats.NextGC.String()

	case "NumForcedGC":
		return servMetrixStats.NumForcedGC.String()

	case "NumGC":
		return servMetrixStats.NumGC.String()

	case "OtherSys":
		return servMetrixStats.OtherSys.String()

	case "PauseTotalNs":
		return servMetrixStats.PauseTotalNs.String()

	case "StackInuse":
		return servMetrixStats.StackInuse.String()

	case "StackSys":
		return servMetrixStats.StackSys.String()

	case "Sys":
		return servMetrixStats.Sys.String()

	case "TotalAlloc":
		return servMetrixStats.TotalAlloc.String()

	case "RandomValue":
		return servMetrixStats.RandomValue.String()

	case "PollCount":
		return servMetrixStats.PollCount.String()
	}

	return "-"
}

func main() {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.HandleFunc("/", handleFunc)
	r.NotFound(notFound)

	servMetrixStats.PollCount = 0

	r.Get("/", func(rw http.ResponseWriter, rq *http.Request) {
		rw.WriteHeader(http.StatusOK)
		textMetricsAndValue := textMetricsAndValue()

		_, err := io.WriteString(rw, textMetricsAndValue)
		if err != nil {
			panic(err)
		}
	})

	r.Get("/value/{metType}/{metName}", func(rw http.ResponseWriter, rq *http.Request) {

		metType := chi.URLParam(rq, "metType")
		metName := chi.URLParam(rq, "metName")

		if metName == "" || metType == "" {
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(getValueMetrics(metType, metName)))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

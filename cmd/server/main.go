package main

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type metric interface {
	Type() string
	String() string
}

var metGauge = map[string]metric{}

var metCounter = map[string]models.Counter{}

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

func handleFunc(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//fmt.Fprintf(w, "err %q\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	message := string(b)
	messageRaz := strings.Split(message, "\n")

	for _, val := range messageRaz {

		typeMetric := valStrMetrics(val, 4)
		nameMetric := valStrMetrics(val, 5)

		switch typeMetric {
		case constants.MetricGauge:
			valueGauge := valueGauge(val)
			metGauge[nameMetric] = models.Gauge(valueGauge)
		case constants.MetricCounter:
			valueCounter := valueCounter(val)
			metCounter[nameMetric] = models.Counter(valueCounter)
		}
	}

	w.WriteHeader(http.StatusOK)
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

	for k, v := range metGauge {
		msg = append(msg, fmt.Sprintf(msgFormat, k, v.String()))
	}
	for k, v := range metCounter {
		msg = append(msg, fmt.Sprintf(msgFormat, k, v.String()))
	}

	return strings.Join(msg, "\n")
}

func getValueMetrics(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	if metName == "" || metType == "" {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	switch metType {
	case constants.MetricGauge:
		if _, ok := metGauge[metName]; !ok {
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
			return
		}

		val := metGauge[metName].String()
		//rw.Write([]byte(val))
		_, err := io.WriteString(rw, val)
		if err != nil {
			panic(err)
		}
	case constants.MetricCounter:
		if _, ok := metCounter[metName]; !ok {
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
			return
		}

		val := metCounter[metName].String()
		//rw.Write([]byte(val))
		_, err := io.WriteString(rw, val)
		if err != nil {
			panic(err)
		}
	}

	rw.WriteHeader(http.StatusOK)
}

func setValueMetricsGET(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if metName == "" || metType == "" || metValue == "" {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	var realMetValue string
	switch metType {
	case constants.MetricGauge:
		realMetValue = metGauge[metName].String()
	case constants.MetricCounter:
		realMetValue = metCounter[metName].String()
	}

	if metValue != realMetValue {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Ожидаемое значенние "+metValue+" метрики "+metName+" с типом "+metType+
			" не найдена", http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func setValueMetricsPOST(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	//if metN/ame == "" || metType == "" || metValue == "" {
	//	http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
	//	return
	//}
	//
	//if metType != "gauge" && metType != "counter" {
	//	http.Error(rw, "Тип "+metType+" не обрабатывается", http.StatusNotImplemented)
	//	return
	//}

	switch metType {
	case constants.MetricGauge:
		fltGauge, err := strconv.ParseFloat(metValue, 64)
		if err != nil {
			http.Error(rw, "Метрику "+metName+" с типом "+metType+" нельзя привести к значению "+metValue,
				http.StatusBadRequest)
			return
		}
		metGauge[metName] = models.Gauge(fltGauge)
	case constants.MetricCounter:
		intCounter, err := strconv.ParseInt(metValue, 10, 64)
		if err != nil {
			http.Error(rw, "Метрику "+metName+" с типом "+metType+" нельзя привести к значению "+metValue,
				http.StatusBadRequest)
			return
		}
		predVal := metCounter[metName]
		metCounter[metName] = predVal.Add(intCounter) // metCounter[metName] + models.Counter(intCounter)
	default:
		http.Error(rw, "Тип "+metType+" не обрабатывается", http.StatusNotImplemented)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func main() {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.HandleFunc("/", handleFunc)
	r.NotFound(notFound)

	metCounter["PollCount"] = 0

	r.Get("/", func(rw http.ResponseWriter, rq *http.Request) {
		textMetricsAndValue := textMetricsAndValue()

		_, err := io.WriteString(rw, textMetricsAndValue)
		if err != nil {
			panic(err)
		}
		rw.WriteHeader(http.StatusOK)
	})

	r.Use(middleware.StripSlashes)

	r.Get("/value/{metType}/{metName}", getValueMetrics)
	r.Get("/value/{metType}/{metName}/", getValueMetrics)

	r.Get("/update/{metType}/{metName}/{metValue}", setValueMetricsGET)
	r.Get("/update/{metType}/{metName}/{metValue}/", setValueMetricsGET)

	r.Post("/update/{metType}/{metName}/{metValue}", setValueMetricsPOST)
	r.Post("/update/{metType}/{metName}/{metValue}/", setValueMetricsPOST)

	log.Fatal(http.ListenAndServe(":8080", r))
}

package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type RepStore struct {
	Repo   repository.MapMetrics
	Router chi.Router
}

type MetricType int

const (
	NotFoundMetric MetricType = iota
	GaugeMetric
	CounterMetric
)

func (mt MetricType) String() string {
	return [...]string{"NotFound", "gauge", "counter"}[mt]
}

func setValueInMapa(mapa repository.MetricsType, metType string, metName string, metValue string) int {
	var gm = GaugeMetric
	var cm = CounterMetric

	switch metType {
	case gm.String():
		predVal, err := strconv.ParseFloat(metValue, 64)
		if err != nil {
			fmt.Println("error convert type")
			return 400
		}
		val := repository.Gauge(predVal)
		val.SetVal(mapa, metName)
	case cm.String():
		predVal, err := strconv.ParseInt(metValue, 10, 64)
		if err != nil {
			fmt.Println("error convert type")
			return 400
		}
		val := repository.Counter(predVal)
		val.SetVal(mapa, metName)
	default:
		return 501
	}
	//if metType == repository.Gauge(0).Type() {
	//	predVal, err := strconv.ParseFloat(metValue, 64)
	//	if err != nil {
	//		fmt.Println("error convert type")
	//		return 400
	//	}
	//	val := repository.Gauge(predVal)
	//	val.SetVal(mapa, metName)
	//} else if metType == repository.Counter(0).Type() {
	//	predVal, err := strconv.ParseInt(metValue, 10, 64)
	//	if err != nil {
	//		fmt.Println("error convert type")
	//		return 400
	//	}
	//	val := repository.Counter(predVal)
	//	val.SetVal(mapa, metName)
	//} else {
	//	return 501
	//}

	return 200
}

func handlerNotFound(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (rp *RepStore) handlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	if _, findKey := rp.Repo[metType]; !findKey {
		rw.WriteHeader(404)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", 404)
		return
	}

	mapa := rp.Repo[metType]
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(404) //Вопрос!!!
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", 404)
		return
	}

	var gm = GaugeMetric
	var cm = CounterMetric

	switch metType {
	case gm.String():
		val := mapa[metName].(repository.Gauge)
		strVal := val.String()
		_, err := io.WriteString(rw, strVal)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	case cm.String():
		val := mapa[metName].(repository.Counter)
		strVal := val.String()
		_, err := io.WriteString(rw, strVal)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	rw.WriteHeader(http.StatusOK)
}

func (rp *RepStore) handlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := rp.Repo[metType]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	mapa := rp.Repo[metType]
	httpStatus := setValueInMapa(mapa, metType, metName, metValue)

	rw.WriteHeader(httpStatus)
}

func (rp *RepStore) handlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := rp.Repo[metType]; !findKey {
		mapa := make(repository.MetricsType)
		rp.Repo[metType] = mapa
	}

	mapa := rp.Repo[metType]
	httpStatus := setValueInMapa(mapa, metType, metName, metValue)

	rw.WriteHeader(httpStatus)
}

func (rp *RepStore) handleFunc(rw http.ResponseWriter, rq *http.Request) {

	rw.WriteHeader(http.StatusOK)
}

func TextMetricsAndValue(rp *RepStore) string {
	const msgFormat = "%s = %s"

	var msg []string
	for _, mapa := range rp.Repo {
		for key, val := range mapa {
			msg = append(msg, fmt.Sprintf(msgFormat, key, val))
		}
	}

	return strings.Join(msg, "\n")
}

func (rp *RepStore) handlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	textMetricsAndValue := TextMetricsAndValue(rp)
	_, err := io.WriteString(rw, textMetricsAndValue)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func CreateServerChi() chi.Router {
	nr := chi.NewRouter()

	nr.Use(middleware.RequestID)
	nr.Use(middleware.RealIP)
	nr.Use(middleware.Logger)
	nr.Use(middleware.Recoverer)
	nr.Use(middleware.StripSlashes)

	metrics := make(repository.MapMetrics)
	rp := &RepStore{metrics, nr}

	nr.HandleFunc("/", rp.handleFunc)
	nr.NotFound(handlerNotFound)

	nr.Get("/", rp.handlerGetAllMetrics)
	nr.Get("/value/{metType}/{metName}", rp.handlerGetValue)
	nr.Get("/update/{metType}/{metName}/{metValue}", rp.handlerSetMetrica)
	nr.Post("/update/{metType}/{metName}/{metValue}", rp.handlerSetMetricaPOST)

	return nr
}

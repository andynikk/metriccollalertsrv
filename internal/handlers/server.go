package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"io"
	"net/http"
	"strconv"
	"sync"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricType int
type MetricError int

type HTMLParam struct {
	Title       string
	TextMetrics []string
}

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

const (
	GaugeMetric MetricType = iota
	CounterMetric

	NotError MetricError = iota
	ErrorConvert
	ErrorGetType
)

type RepStore struct {
	MutexRepo repository.MapMetrics
	Router    chi.Router
	mx        sync.Mutex
}

func (rs *RepStore) New() {

	rs.MutexRepo = make(repository.MapMetrics)
	rs.Router = chi.NewRouter()

	rs.Router.Use(middleware.RequestID)
	rs.Router.Use(middleware.RealIP)
	rs.Router.Use(middleware.Logger)
	rs.Router.Use(middleware.Recoverer)
	rs.Router.Use(middleware.StripSlashes)

	rs.Router.HandleFunc("/", rs.HandleFunc)
	rs.Router.NotFound(HandlerNotFound)

	rs.Router.Get("/", rs.HandlerGetAllMetrics)
	rs.Router.Get("/value/{metType}/{metName}", rs.HandlerGetValue)
	rs.Router.Get("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetrica)
	rs.Router.Post("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetricaPOST)
	rs.Router.Post("/update", rs.HandlerUpdateMetricaJSON)
	rs.Router.Post("/value", rs.HandlerValueMetricaJSON)

}

func (rs *RepStore) GetCounter(tm string, key string) repository.Counter {

	return rs.MutexRepo[tm][key].(repository.Counter)

}

func (rs *RepStore) SetCounter(tm string, key string, value repository.Counter) {
	rs.mx.Lock()
	defer rs.mx.Unlock()

	if _, findKey := rs.MutexRepo[tm][key]; !findKey {
		rs.MutexRepo[tm][key] = value
	} else {
		rs.MutexRepo[tm][key] = rs.MutexRepo[tm][key].(repository.Counter) + value
	}
}

func (rs *RepStore) SetGauge(tm string, key string, value repository.Gauge) {
	rs.mx.Lock()
	defer rs.mx.Unlock()

	rs.MutexRepo[tm][key] = value

}

func (rs *RepStore) setValueInMapa(metType string, metName string, metValue string) MetricError {
	var gm = GaugeMetric
	var cm = CounterMetric

	var ne = NotError
	var ec = ErrorConvert
	var egt = ErrorGetType

	switch metType {
	case gm.String():
		predVal, err := strconv.ParseFloat(metValue, 64)
		if err != nil {
			fmt.Println("error convert type")
			return ec
		}
		val := repository.Gauge(predVal)
		rs.SetGauge(metType, metName, val)

	case cm.String():
		predVal, err := strconv.ParseInt(metValue, 10, 64)
		if err != nil {
			return ec
		}
		val := repository.Counter(predVal)
		rs.SetCounter(metType, metName, val)
	default:
		return egt
	}

	return ne
}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	rs.mx.Lock()
	defer rs.mx.Unlock()

	if _, findKey := rs.MutexRepo[metType]; !findKey {
		mapa := make(repository.MetricsType)
		rs.MutexRepo[metType] = mapa
	}

	mapa := rs.MutexRepo[metType]
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
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

func (rs *RepStore) HandlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := rs.MutexRepo[metType]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	var ec = ErrorConvert
	var egt = ErrorGetType

	errStatus := rs.setValueInMapa(metType, metName, metValue)
	switch errStatus {
	case egt:
		rw.WriteHeader(http.StatusNotImplemented)
	case ec:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}

}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := rs.MutexRepo[metType]; !findKey {

		rs.mx.Lock()

		mapa := make(repository.MetricsType)
		rs.MutexRepo[metType] = mapa

		rs.mx.Unlock()
	}

	var ec = ErrorConvert
	var egt = ErrorGetType

	errStatus := rs.setValueInMapa(metType, metName, metValue)
	switch errStatus {
	case egt:
		rw.WriteHeader(http.StatusNotImplemented)
	case ec:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}
}

func (rs *RepStore) HandlerUpdateMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	v := encoding.Metrics{}
	err := json.NewDecoder(rq.Body).Decode(&v)
	if err != nil {
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	if _, findKey := rs.MutexRepo[metType]; !findKey {
		rs.mx.Lock()

		mapa := make(repository.MetricsType)
		rs.MutexRepo[metType] = mapa

		rs.mx.Unlock()
	}

	var gm = GaugeMetric
	var cm = CounterMetric

	switch metType {
	case gm.String():
		rs.SetGauge(metType, metName, v.Gauge())
	case cm.String():
		rs.SetCounter(metType, metName, v.Counter())
	default:
		rw.WriteHeader(http.StatusNotImplemented)
		return
	}

	//var ec = ErrorConvert
	//var egt = ErrorGetType
	//
	//errStatus := rs.setValueInMapa(metType, metName, metValue)
	//switch errStatus {
	//case egt:
	//	rw.WriteHeader(http.StatusNotImplemented)
	//case ec:
	//	rw.WriteHeader(http.StatusBadRequest)
	//default:
	//	rw.WriteHeader(http.StatusOK)
	//}
}

func (rs *RepStore) HandlerValueMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	v := encoding.Metrics{}
	err := json.NewDecoder(rq.Body).Decode(&v)
	if err != nil {
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	rs.mx.Lock()
	defer rs.mx.Unlock()

	if _, findKey := rs.MutexRepo[metType]; !findKey {
		mapa := make(repository.MetricsType)
		rs.MutexRepo[metType] = mapa
	}

	mapa := rs.MutexRepo[metType]
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	var gm = GaugeMetric
	var cm = CounterMetric

	switch metType {
	case gm.String():

		val := mapa[metName].(repository.Gauge).Float64()

		mt := encoding.Metrics{ID: metName, MType: metType, Value: &val}
		arrJSON, err := mt.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		bt := bytes.NewBuffer(arrJSON).String()
		_, err = io.WriteString(rw, bt)
		rw.Write(arrJSON)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("result", bt)

		rq.Header.Set("Content-Type", "application/json")
		rq.Header.Set("result", bt)

	case cm.String():
		val := mapa[metName].(repository.Counter).Int64()

		mt := encoding.Metrics{ID: metType, MType: metName, Delta: &val}
		arrJSON, err := mt.MarshalMetrica()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		bt := bytes.NewBuffer(arrJSON).String()
		_, err = io.WriteString(rw, bt)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		rq.Header.Set("Content-Type", "application/json")
		rq.Header.Set("result", bt)

	default:
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()

	arrMetricsAndValue := textMetricsAndValue(rs.MutexRepo)

	data := HTMLParam{
		Title:       "МЕТРИКИ",
		TextMetrics: arrMetricsAndValue,
	}

	tmpl, errTpl := template.ParseFiles("internal/templates/home_pages.html")
	if errTpl != nil {
		http.Error(rw, errTpl.Error(), http.StatusServiceUnavailable)
		return
	}
	tmpl.Execute(rw, data)

	rw.WriteHeader(http.StatusOK)
}

func HandlerNotFound(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
}

func textMetricsAndValue(mm repository.MapMetrics) []string {
	const msgFormat = "%s = %s"

	var msg []string

	for _, mapa := range mm {
		for key, val := range mapa {
			msg = append(msg, fmt.Sprintf(msgFormat, key, val))
		}
	}

	return msg //strings.Join(msg, "<br />")
}

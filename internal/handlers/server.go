package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"sync"
	"text/template"

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

type Config struct {
	STORE_INTERVAL int64  `env:"STORE_INTERVAL" envDefault:"300"`
	STORE_FILE     string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	RESTORE        bool   `env:"RESTORE" envDefault:"true"`
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
	rs.Router.Post("/update", rs.HandlerUpdateMetricJSON)
	rs.Router.Post("/value", rs.HandlerValueMetricaJSON)

}

func (rs *RepStore) addNilMetric(metType string, metName string) MetricError {
	var GaugeMetric = GaugeMetric
	var CounterMetric = CounterMetric

	switch metType {
	case GaugeMetric.String():
		var nilGauge *repository.Gauge
		rs.MutexRepo[metName] = nilGauge

		fl := repository.Gauge(0)
		valGauge := &fl

		rs.MutexRepo[metName] = valGauge
	case CounterMetric.String():
		var nilCounter *repository.Counter
		rs.MutexRepo[metName] = nilCounter

		in := repository.Counter(0)
		valCounter := &in

		rs.MutexRepo[metName] = valCounter
	default:
		return ErrorGetType
	}

	return NotError
}

func (rs *RepStore) setValueInMapa(metType string, metName string, metValue string) MetricError {

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		status := rs.addNilMetric(metType, metName)
		if status != NotError {
			return status
		}
	}

	status := rs.MutexRepo[metName].SetFromText(metValue)

	switch status {
	case 1:
		return ErrorConvert
	case 0:
		return NotError
	}

	return NotError

}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	rs.mx.Lock()
	defer rs.mx.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	strMetric := rs.MutexRepo[metName].String()
	_, err := io.WriteString(rw, strMetric)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	errStatus := NotError
	if _, findKey := rs.MutexRepo[metName]; !findKey {
		errStatus = rs.addNilMetric(metType, metName)
	}

	if errStatus == NotError {
		errStatusInt := rs.MutexRepo[metName].SetFromText(metValue)
		if errStatusInt == 1 {
			errStatus = ErrorConvert
		}
	}

	switch errStatus {
	case ErrorGetType:
		rw.WriteHeader(http.StatusNotImplemented)
	case 1:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}

}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

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

func (rs *RepStore) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	v := encoding.Metrics{}
	err := json.NewDecoder(rq.Body).Decode(&v)

	if err != nil {
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rs.addNilMetric(metType, metName)
	}

	//mt := rs.MutexRepo[metName].GetMetrics(metName, metType)
	rs.MutexRepo[metName].Set(v)
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

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName)
	metricsJSON, err := mt.MarshalMetrica()

	rw.Header().Add("Content-Type", "application/json")
	if _, err := rw.Write(metricsJSON); err != nil {
		fmt.Println(err.Error())
		return
	}
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

	for key, val := range mm {
		msg = append(msg, fmt.Sprintf(msgFormat, key, val.String()))
	}

	return msg
}

func JSONMetricsAndValue(mm repository.MapMetrics) []encoding.Metrics {

	var arr []encoding.Metrics

	for key, val := range mm {
		jMetric := val.GetMetrics(val.Type(), key)
		arr = append(arr, jMetric)
	}

	return arr
}

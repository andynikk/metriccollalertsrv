package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v6"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
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
	Router    chi.Router
	MX        sync.Mutex
	MutexRepo repository.MapMetrics
}

type Config struct {
	STORE_INTERVAL int64  `env:"STORE_INTERVAL" envDefault:"300"`
	STORE_FILE     string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	RESTORE        bool   `env:"RESTORE" envDefault:"true"`
	ADDRESS        string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func NewRepStore() *RepStore {

	rp := new(RepStore)
	rp.New()

	return rp
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

func (rs *RepStore) AddNilMetric(metType string, metName string) MetricError {
	var GaugeMetric = GaugeMetric
	var CounterMetric = CounterMetric

	fmt.Println(metType, metName)

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

	rs.MX.Lock()
	defer rs.MX.Unlock()

	fmt.Println("metType 2", metType)
	if _, findKey := rs.MutexRepo[metName]; !findKey {
		fmt.Println("metType 3", metType)
		status := rs.AddNilMetric(metType, metName)
		if status != NotError {
			fmt.Println("metType 4", metType)
			return status
		}
	}

	fmt.Println("metType 5", metType)
	status := rs.MutexRepo[metName].SetFromText(metValue)
	return status
}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	rs.MX.Lock()
	defer rs.MX.Unlock()

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
	rs.MX.Lock()
	defer rs.MX.Unlock()

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
		errStatus = rs.AddNilMetric(metType, metName)
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
	case ErrorConvert:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}
}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")
	//if metType == "unknown" {
	//	rw.WriteHeader(http.StatusNotImplemented)
	//	return
	//}

	fmt.Println("metType 1")
	errStatus := rs.setValueInMapa(metType, metName, metValue)
	fmt.Println(metType, metName, metValue, errStatus)

	switch errStatus {
	case ErrorGetType:
		rw.WriteHeader(http.StatusNotImplemented)
	case ErrorConvert:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}
}

func (rs *RepStore) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	rs.MX.Lock()
	defer rs.MX.Unlock()

	v := encoding.Metrics{}
	err := json.NewDecoder(rq.Body).Decode(&v)

	if err != nil {
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rs.AddNilMetric(metType, metName)
	}
	rs.MutexRepo[metName].Set(v)

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	if cfg.STORE_INTERVAL == 0 {
		patch := cfg.STORE_FILE
		rs.SaveMetric2File(patch)
	}
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

	rs.MX.Lock()
	defer rs.MX.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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

func (rs *RepStore) SaveMetric2File(patch string) {

	arr := JSONMetricsAndValue(rs.MutexRepo)
	arrJSON, err := json.Marshal(arr)
	if err != nil {
		fmt.Println(err.Error())
	}
	if patch == "" {
		return
	}

	if err := ioutil.WriteFile(patch, arrJSON, 0777); err != nil {
		fmt.Println(err.Error())
	}

}

func HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Метрика "+r.URL.Path+" не найдена", http.StatusNotFound)

	//_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	//if err != nil {
	//	http.Error(rw, err.Error(), http.StatusNotFound)
	//	return
	//}
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
		jMetric := val.GetMetrics(key, val.Type())
		arr = append(arr, jMetric)
	}

	return arr
}

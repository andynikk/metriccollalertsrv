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
	//rs.Router.Get("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetrica)
	rs.Router.Post("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetricaPOST)
	rs.Router.Post("/update", rs.HandlerUpdateMetricJSON)
	rs.Router.Post("/value", rs.HandlerValueMetricaJSON)

}

func (rs *RepStore) setValueInMap(metType string, metName string, metValue string) MetricError {

	//rs.MX.Lock()
	//defer rs.MX.Unlock()

	switch metType {
	case GaugeMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if err := val.SetFromText(metValue); err == 400 {
				return http.StatusBadRequest
			}
		} else {

			valG := repository.Gauge(0)
			if err := valG.SetFromText(metValue); err == 400 {
				return http.StatusBadRequest
			}

			rs.MutexRepo[metName] = &valG
		}

	case CounterMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if err := val.SetFromText(metValue); err == 400 {
				return http.StatusBadRequest
			}
		} else {

			valC := repository.Counter(0)
			if err := valC.SetFromText(metValue); err == 400 {
				return http.StatusBadRequest
			}

			rs.MutexRepo[metName] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	return http.StatusOK
}

func (rs *RepStore) SetValueInMapJSON(v encoding.Metrics) MetricError {

	switch v.MType {
	case GaugeMetric.String():
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valG := repository.Gauge(0)
			rs.MutexRepo[v.ID] = &valG
		}
	case CounterMetric.String():
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valC := repository.Counter(0)
			rs.MutexRepo[v.ID] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	rs.MutexRepo[v.ID].Set(v)
	return http.StatusOK
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

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	rs.MX.Lock()
	defer rs.MX.Unlock()

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	errStatus := rs.setValueInMap(metType, metName, metValue)

	switch errStatus {
	case 400:
		rw.WriteHeader(http.StatusBadRequest)
	case 501:
		rw.WriteHeader(http.StatusNotImplemented)
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

	fmt.Println("Пришла метрика", v.ID, v.MType, v.Value, v.Delta)
	errStatus := rs.SetValueInMapJSON(v)
	fmt.Println("Статус установки значений метрики", errStatus)

	switch errStatus {
	case 400:
		rw.WriteHeader(http.StatusBadRequest)
	case 501:
		rw.WriteHeader(http.StatusNotImplemented)
	default:
		rw.WriteHeader(http.StatusOK)
	}

	mt := rs.MutexRepo[v.ID].GetMetrics(v.MType, v.ID)
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

	//b, err := ioutil.ReadAll(rq.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("Пришл текст: %s", b)
	fmt.Printf("Количество метрик: %d\n", len((rs.MutexRepo)))

	v := encoding.Metrics{}
	if err := json.NewDecoder(rq.Body).Decode(&v); err != nil {
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}

	metType := v.MType
	metName := v.ID

	fmt.Println("Пришла метрика:", v.MType, v.ID)

	rs.MX.Lock()
	defer rs.MX.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		fmt.Println("Метрика не найдена:", v.MType, v.ID)

		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		fmt.Println("Метрика не получена:", v.MType, v.ID)
		fmt.Println(err.Error())
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	if _, err := rw.Write(metricsJSON); err != nil {
		fmt.Println("Метрика не вписано в тело:", v.MType, v.ID)
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
	rw.WriteHeader(http.StatusNotFound)

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

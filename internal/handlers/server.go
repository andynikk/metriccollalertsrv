package handlers

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"strconv"
	"sync"
	"text/template"
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

}

func (c *RepStore) GetCounter(tm string, key string) repository.Counter {

	return c.MutexRepo[tm][key].(repository.Counter)

}

func (c *RepStore) SetCounter(tm string, key string, value repository.Counter) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, findKey := c.MutexRepo[tm][key]; !findKey {
		c.MutexRepo[tm][key] = value
	} else {
		c.MutexRepo[tm][key] = c.MutexRepo[tm][key].(repository.Counter) + value
	}
}

func (c *RepStore) SetGauge(tm string, key string, value repository.Gauge) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.MutexRepo[tm][key] = value

}

func (mp *RepStore) setValueInMapa(metType string, metName string, metValue string) MetricError {
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
		mp.SetGauge(metType, metName, val)

	case cm.String():
		predVal, err := strconv.ParseInt(metValue, 10, 64)
		if err != nil {
			return ec
		}
		val := repository.Counter(predVal)
		mp.SetCounter(metType, metName, val)
	default:
		return egt
	}

	return ne
}

func (c *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	c.mx.Lock()
	defer c.mx.Unlock()

	if _, findKey := c.MutexRepo[metType]; !findKey {

		mapa := make(repository.MetricsType)
		c.MutexRepo[metType] = mapa
	}

	mapa := c.MutexRepo[metType]
	//i := 0
	//for _, k := range mapa {
	//	fmt.Println(k)
	//	i++
	//}
	//if i == 0 {
	//	rw.WriteHeader(http.StatusNotImplemented)
	//	http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotImplemented)
	//	return
	//}
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
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

func (c *RepStore) HandlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := c.MutexRepo[metType]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	var ec = ErrorConvert
	var egt = ErrorGetType

	errStatus := c.setValueInMapa(metType, metName, metValue)
	switch errStatus {
	case egt:
		rw.WriteHeader(http.StatusNotImplemented)
	case ec:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}

}

func (c *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := c.MutexRepo[metType]; !findKey {

		c.mx.Lock()

		mapa := make(repository.MetricsType)
		c.MutexRepo[metType] = mapa

		c.mx.Unlock()
	}

	var ec = ErrorConvert
	var egt = ErrorGetType

	errStatus := c.setValueInMapa(metType, metName, metValue)
	switch errStatus {
	case egt:
		rw.WriteHeader(http.StatusNotImplemented)
	case ec:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}
}

func (c *RepStore) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (c *RepStore) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()

	arrMetricsAndValue := textMetricsAndValue(c.MutexRepo)

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

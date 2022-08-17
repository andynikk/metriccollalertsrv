package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type RepStore struct {
	MutexRepo repository.MapMetrics
	Router    chi.Router
	mx        sync.Mutex
}

type MetricType int
type MetricError int

const (
	GaugeMetric MetricType = iota
	CounterMetric

	NotError MetricError = iota
	ErrorConvert
	ErrorGetType
)

func (c *RepStore) SetCounter(tm string, key string, value repository.Counter) {
	c.mx.Lock()

	if _, findKey := c.MutexRepo[tm][key]; !findKey {
		c.MutexRepo[tm][key] = value

	} else {
		c.MutexRepo[tm][key] = c.MutexRepo[tm][key].(repository.Counter) + value
	}
	c.mx.Unlock()
}

func (c *RepStore) SetGauge(tm string, key string, value repository.Gauge) {
	c.mx.Lock()

	c.MutexRepo[tm][key] = value

	c.mx.Unlock()
}

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

func setValueInMapa(mp *RepStore, metType string, metName string, metValue string) MetricError {
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
			//400
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
		//val.SetVal(mapa, metName)
	default:
		return egt
	}

	return ne
}

func handlerNotFound(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
}

func (c *RepStore) handlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	if _, findKey := c.MutexRepo[metType]; !findKey {

		mapa := make(repository.MetricsType)
		c.MutexRepo[metType] = mapa
	}

	mapa := c.MutexRepo[metType]
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(http.StatusNotFound) //Вопрос!!!
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

func (c *RepStore) handlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
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

	errStatus := setValueInMapa(c, metType, metName, metValue)
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

		mapa := make(repository.MetricsType)
		c.MutexRepo[metType] = mapa
	}

	var ec = ErrorConvert
	var egt = ErrorGetType

	errStatus := setValueInMapa(c, metType, metName, metValue)
	switch errStatus {
	case egt:
		rw.WriteHeader(http.StatusNotImplemented)
	case ec:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusOK)
	}
}

func (c *RepStore) handleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func TextMetricsAndValue(rp *RepStore) []string {
	const msgFormat = "%s = %s"

	var msg []string

	for _, mapa := range rp.MutexRepo {
		for key, val := range mapa {
			msg = append(msg, fmt.Sprintf(msgFormat, key, val))
		}
	}

	return msg //strings.Join(msg, "<br />")
}

type HTMLParam struct {
	Title       string
	TextMetrics []string
}

func (c *RepStore) handlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()

	arrMetricsAndValue := TextMetricsAndValue(c)

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

func CreateServerChi() chi.Router {
	nr := chi.NewRouter()

	nr.Use(middleware.RequestID)
	nr.Use(middleware.RealIP)
	nr.Use(middleware.Logger)
	nr.Use(middleware.Recoverer)
	nr.Use(middleware.StripSlashes)

	//mt := make(repository.MetricsType)
	//mtm := repository.MutexTypeMetrics{M: mt}

	mm := make(repository.MapMetrics)
	rp := &RepStore{MutexRepo: mm, Router: nr}

	nr.HandleFunc("/", rp.handleFunc)
	nr.NotFound(handlerNotFound)

	nr.Get("/", rp.handlerGetAllMetrics)
	nr.Get("/value/{metType}/{metName}", rp.handlerGetValue)
	nr.Get("/update/{metType}/{metName}/{metValue}", rp.handlerSetMetrica)
	nr.Post("/update/{metType}/{metName}/{metValue}", rp.HandlerSetMetricaPOST)

	return nr
}

package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

func setValueInMapa(mapa repository.MetricsType, metType string, metName string, metValue string) int {
	if metType == repository.Gauge(0).Type() {
		predVal, err := strconv.ParseFloat(metValue, 64)
		if err != nil {
			fmt.Println("error convert type")
			return 400
		}
		val := repository.Gauge(predVal)
		val.SetVal(mapa, metName)
	} else if metType == repository.Counter(0).Type() {
		predVal, err := strconv.ParseInt(metValue, 10, 64)
		if err != nil {
			fmt.Println("error convert type")
			return 400
		}
		val := repository.Counter(predVal)
		val.SetVal(mapa, metName)
	} else {
		return 501
	}

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

func handlerGetValue(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	if _, findKey := repository.Metrics[metType]; !findKey {
		rw.WriteHeader(404)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", 404)
		return
	}

	mapa := repository.Metrics[metType]
	if _, findKey := mapa[metName]; !findKey {
		rw.WriteHeader(404) //Вопрос!!!
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", 404)
		return
	}

	if metType == "gauge" {
		val := mapa[metName].(repository.Gauge)
		strVal := val.String()
		_, err := io.WriteString(rw, strVal)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	} else {
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

func handlerSetMetrica(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := repository.Metrics[metType]; !findKey {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	mapa := repository.Metrics[metType]
	httpStatus := setValueInMapa(mapa, metType, metName, metValue)

	rw.WriteHeader(httpStatus)
}

func handlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if _, findKey := repository.Metrics[metType]; !findKey {
		mapa := make(repository.MetricsType)
		repository.Metrics[metType] = mapa
	}

	mapa := repository.Metrics[metType]
	httpStatus := setValueInMapa(mapa, metType, metName, metValue)

	rw.WriteHeader(httpStatus)
}

func handleFunc(rw http.ResponseWriter, rq *http.Request) {

	rw.WriteHeader(http.StatusOK)
}

func handlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	textMetricsAndValue := repository.TextMetricsAndValue()

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

	nr.HandleFunc("/", handleFunc)
	nr.NotFound(handlerNotFound)

	nr.Get("/", handlerGetAllMetrics)
	nr.Get("/value/{metType}/{metName}", handlerGetValue)
	nr.Get("/update/{metType}/{metName}/{metValue}", handlerSetMetrica)
	nr.Post("/update/{metType}/{metName}/{metValue}", handlerSetMetricaPOST)

	return nr
}

package main

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func valStrMetrics(strMetrix string, numElArr int64) string {
	arrMetrics := strings.Split(strMetrix, "/")
	valMetrics := arrMetrics[numElArr]

	return valMetrics
}

func textMetricsAndValue() string {
	const msgFormat = "%s = %s"

	var msg []string

	for key, val := range repository.Metrics {
		msg = append(msg, fmt.Sprintf(msgFormat, key, val))
	}

	return strings.Join(msg, "\n")
}

func HandleFunc(w http.ResponseWriter, r *http.Request) {

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
		valMetric := valStrMetrics(val, 6)

		repository.SetValue(typeMetric, nameMetric, valMetric)
	}

	w.WriteHeader(http.StatusOK)
}

func main() {

	nr := chi.NewRouter()

	nr.Use(middleware.RequestID)
	nr.Use(middleware.RealIP)
	nr.Use(middleware.Logger)
	nr.Use(middleware.Recoverer)
	nr.Use(middleware.StripSlashes)

	nr.HandleFunc("/", HandleFunc)
	nr.NotFound(handlers.NotFound)

	repository.RefTypeMepStruc()
	repository.Metrics["PollCount"] = 0

	nr.Get("/", func(rw http.ResponseWriter, rq *http.Request) {
		textMetricsAndValue := textMetricsAndValue()

		_, err := io.WriteString(rw, textMetricsAndValue)
		if err != nil {
			log.Fatal(err)
			return
		}
		rw.WriteHeader(http.StatusOK)
	})

	nr.Get("/value/{metType}/{metName}", handlers.GetValueHandler)
	nr.Get("/update/{metType}/{metName}/{metValue}", handlers.SetValueGETHandler)
	nr.Post("/update/{metType}/{metName}/{metValue}", handlers.SetValuePOSTHandler)

	log.Fatal(http.ListenAndServe(":8080", nr))
}

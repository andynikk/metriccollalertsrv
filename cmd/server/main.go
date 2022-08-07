package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

func (g gauge) Type() string {
	return fmt.Sprintf("%T", g)
}

func (c counter) Type() string {
	return fmt.Sprintf("%T", c)
}

func (c counter) String() string {
	return fmt.Sprintf("%d", int64(c))
}

func (g gauge) String() string {
	fg := float64(g)
	return fmt.Sprintf("%g", fg)
}

var metGauge = map[string]gauge{}

var metCounter = map[string]counter{}

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

		//typeMetric := valStrMetrics(val, 1)
		//nameMetric := valStrMetrics(val, 2)

		if typeMetric == "gauge" {
			valueGauge := valueGauge(val)
			metGauge[nameMetric] = gauge(valueGauge)
		} else if typeMetric == "counter" {
			valueCounter := valueCounter(val)
			metCounter[nameMetric] = counter(valueCounter)
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

	r.Get("/value/{metType}/{metName}", func(rw http.ResponseWriter, rq *http.Request) {

		metType := chi.URLParam(rq, "metType")
		metName := chi.URLParam(rq, "metName")

		if metName == "" || metType == "" {
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
			return
		}

		rw.WriteHeader(http.StatusOK)
		if metType == "gauge" {
			rw.Write([]byte(metGauge[metName].String()))
			_, err := io.WriteString(rw, metGauge[metName].String())
			if err != nil {
				panic(err)
			}
		} else if metType == "counter" {
			rw.Write([]byte(metCounter[metName].String()))
			_, err := io.WriteString(rw, metCounter[metName].String())
			if err != nil {
				panic(err)
			}
		}
	})

	r.Get("/update/{metType}/{metName}/{metValue}", func(rw http.ResponseWriter, rq *http.Request) {

		metType := chi.URLParam(rq, "metType")
		metName := chi.URLParam(rq, "metName")
		metValue := chi.URLParam(rq, "metValue")

		if metName == "" || metType == "" || metValue == "" {
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
			return
		}

		var realMetValue string
		if metType == "gauge" {
			realMetValue = metGauge[metName].String()
		} else if metType == "counter" {
			realMetValue = metCounter[metName].String()
		}
		if metValue != realMetValue {
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, "Ожидаемое значенние "+metValue+" метрики "+metName+" с типом "+metType+
				" не найдена", http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
	})

	r.Post("/update/{metType}/{metName}/{metValue}", func(rw http.ResponseWriter, rq *http.Request) {

		metType := chi.URLParam(rq, "metType")
		metName := chi.URLParam(rq, "metName")
		metValue := chi.URLParam(rq, "metValue")

		if metName == "" || metType == "" || metValue == "" {
			http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
			return
		}

		if metType != "gauge" && metType != "counter" {
			http.Error(rw, "Тип "+metType+" не обрабатывается", http.StatusNotImplemented)
			return
		}

		if metType == "gauge" {

			fltGauge, err := strconv.ParseFloat(metValue, 64)
			if err != nil {
				http.Error(rw, "Метрику "+metName+" с типом "+metType+" нельзя привести к значению "+metValue,
					http.StatusBadRequest)
				return
			}
			metGauge[metName] = gauge(fltGauge)

		} else if metType == "counter" {

			intCounter, err := strconv.ParseInt(metValue, 10, 64)
			if err != nil {
				http.Error(rw, "Метрику "+metName+" с типом "+metType+" нельзя привести к значению "+metValue,
					http.StatusBadRequest)
				return
			}
			metCounter[metName] = counter(intCounter)
		}

		rw.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

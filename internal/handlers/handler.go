package handlers

import (
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"strconv"
)

func NotFound(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	if r.Method != "GET" {
		return
	}

	_, err := io.WriteString(rw, "Метрика "+r.URL.Path+" не найдена")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetValueHandler(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	if metName == "" || metType == "" {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	if _, findKey := repository.Metrics[metName]; !findKey {
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	val := repository.StringValue(metType, metName)

	_, err := io.WriteString(rw, val)
	if err != nil {
		log.Fatal(err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func SetValueGETHandler(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	if metName == "" || metType == "" || metValue == "" {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusBadRequest)
		return
	}

	realMetValue := repository.StringValue(metType, metName)

	if metValue != realMetValue {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Ожидаемое значенние "+metValue+" метрики "+metName+" с типом "+metType+
			" не найдена", http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func SetValuePOSTHandler(rw http.ResponseWriter, rq *http.Request) {
	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	err := repository.SetValue(metType, metName, metValue)
	if err != nil {
		var val, errConvert = strconv.ParseInt(err.Error(), 10, 64)
		if errConvert != nil {
			http.Error(rw, "Ошибка получения значения ответа http", 501)
		}
		http.Error(rw, "Ошибка установки значения "+metValue+" метрики "+metName+" с типом "+metType,
			int(val))
	}

	rw.WriteHeader(http.StatusOK)
}

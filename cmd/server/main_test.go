package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

func TestFuncServer(t *testing.T) {

	var postStr = "http://127.0.0.1:8080/update/gauge/Alloc/0.1\nhttp://127.0.0.1:8080/update/gauge/" +
		"BuckHashSys/0.002\nhttp://127.0.0.1:8080/update/counter/PollCount/5"

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		messageRaz := strings.Split(postStr, "\n")
		if len(messageRaz) != 3 {
			t.Errorf("The string (%s) was incorrectly decomposed into an array", postStr)
		}
	})

	t.Run("Checking the filling of metrics", func(t *testing.T) {
		t.Run("Checking the type of the first line", func(t *testing.T) {
			var typeGauge = "gauge"

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]

			if strings.Contains(valElArr, typeGauge) == false {
				t.Errorf("The Gauge type was incorrectly determined")
			}
		})

		tests := []struct {
			name           string
			request        string
			wantStatusCode int
		}{
			{name: "Проверка на установку значения counter", request: "/update/counter/testSetGet332/6",
				wantStatusCode: http.StatusOK},
			{name: "Проверка на не правильный тип метрики", request: "/update/notcounter/testSetGet332/6",
				wantStatusCode: http.StatusNotImplemented},
			{name: "Проверка на не правильное значение метрики", request: "/update/counter/testSetGet332/non",
				wantStatusCode: http.StatusBadRequest},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				r := chi.NewRouter()
				ts := httptest.NewServer(r)

				mm := new(repository.StoreMetrics)
				rp := handlers.RepStore{MutexRepo: *mm, Router: nil}
				r.Post("/update/{metType}/{metName}/{metValue}", rp.HandlerSetMetricaPOST)

				defer ts.Close()
				resp := testRequest(t, ts, http.MethodPost, tt.request, nil)
				defer resp.Body.Close()

				if resp.StatusCode != tt.wantStatusCode {
					t.Errorf("Ответ не верен")
				}
			})
		}
	})

	t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
		t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
			var typeCounter = "counter"

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[2]

			if strings.Contains(valElArr, typeCounter) == false {
				t.Errorf("The Counter type was incorrectly determined")
			}
		})

	})

}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	//respBody, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	t.Fatal(err)
	//	return nil
	//}
	defer resp.Body.Close()

	return resp
}

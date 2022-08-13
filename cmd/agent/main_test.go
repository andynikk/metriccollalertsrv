package main

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestmakeMsg(memStats Metrics) string {

	const adresServer = "127.0.0.1:8080"
	const msgFormat = "http://%s/update/%s/%s/%v"

	var msg []string

	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats["Alloc"].Type(), "Alloc", 0.1))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats["BuckHashSys"].Type(), "BuckHashSys", 0.002))

	return strings.Join(msg, "\n")
}

func TestFuncAgen(t *testing.T) {
	var resultMS Metrics
	var argErr = "err"

	t.Run("Checking the structure creation", func(t *testing.T) {

		var realResult Metrics

		if resultMS["Alloc"] != realResult["Alloc"] && resultMS["RandomValue"] != realResult["RandomValue"] {

			//t.Errorf("Structure creation error", resultMS, realResult)
			t.Errorf("Structure creation error (%s)", argErr)
		}
		t.Run("Creating a submission line", func(t *testing.T) {
			var resultStr = "http://127.0.0.1:8080/update/gauge/Alloc/0.1\nhttp://127.0.0.1:8080/update/gauge/BuckHashSys/0.002"

			resultMassage := TestmakeMsg(realResult)

			if resultStr != resultMassage {

				//t.Errorf("Error creating a submission line", string(resultMS), realResult)
				t.Errorf("Error creating a submission line (%s)", argErr)
			}

			//t.Run("Creating a submission line", func(t *testing.T) {
			//
			//	r := strings.NewReader(resultMassage)
			//	resp, err := http.Post("http://localhost:8080", "text/plain", r)
			//
			//	if err != nil {
			//		t.Errorf("Error sending a POST message (%s)", err.Error())
			//	}
			//
			//	if resp.Status != "200 OK" {
			//		t.Errorf("Incorrect jndtnf status (%s)", err.Error())
			//	}
			//	resp.Body.Close()
			//})
		})
	})

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fillMetric(resultMS, &mem)
	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		if resultMS["Frees"].Type() != "gauge" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Gauge")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if resultMS["Alloc"] == 0 {
			t.Errorf("The metric %s a value of %v", "Alloc", 0)
		}

	})

	fillMetric(resultMS, &mem)
	t.Run("Checking the filling of metrics PollCount", func(t *testing.T) {

		if resultMS["PollCount"].Type() != "counter" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Counter")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if resultMS["PollCount"] == 0 {
			t.Errorf("The metric %s a value of %v", "PollCount", 0)
		}

	})

	t.Run("Increasing the metric PollCount", func(t *testing.T) {
		var res = int64(1)

		if PollCount != res {
			t.Errorf("The metric %s has not increased by %v", "PollCount", res)
		}

	})

}

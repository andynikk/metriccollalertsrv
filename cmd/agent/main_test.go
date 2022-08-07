package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

func TestmakeMsg(memStats MemStats) string {

	const adresServer = "127.0.0.1:8080"
	const msgFormat = "http://%s/update/%s/%s/%v"

	var msg []string

	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.Alloc.Type(), "Alloc", 0.1))
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, memStats.BuckHashSys.Type(), "BuckHashSys", 0.002))

	return strings.Join(msg, "\n")
}

func TestFuncAgen(t *testing.T) {
	var resultMS = MemStats{}
	var argErr = "err"

	t.Run("Checking the structure creation", func(t *testing.T) {

		realResult := MemStats{}

		if resultMS.Alloc != realResult.Alloc && resultMS.RandomValue != realResult.RandomValue {

			//t.Errorf("Structure creation error", resultMS, realResult)
			t.Errorf("Structure creation error (%s)", argErr)
		}
		t.Run("Creating a submission line", func(t *testing.T) {
			var resultStr = "http://127.0.0.1:8080/update/main.gauge/Alloc/0.1\nhttp://127.0.0.1:8080/update/main.gauge/BuckHashSys/0.002"

			resultMassage := TestmakeMsg(realResult)

			if resultStr != resultMassage {

				//t.Errorf("Error creating a submission line", string(resultMS), realResult)
				t.Errorf("Error creating a submission line (%s)", argErr)
			}

			t.Run("Creating a submission line", func(t *testing.T) {

				resp, err := http.Post("http://127.0.0.1:8080", "text/plain", strings.NewReader(resultMassage))
				defer resp.Body.Close()

				if err != nil {
					t.Errorf("Error sending a POST message (%s)", err.Error())
				}

				if resp.Status != "200 OK" {
					t.Errorf("Incorrect jndtnf status (%s)", err.Error())
				}
			})
		})
	})

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fillGauge(&resultMS, &mem)
	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		if resultMS.Frees.Type() != "main.gauge" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Gauge")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if resultMS.Alloc == 0 {
			t.Errorf("The metric %s a value of %v", "Alloc", 0)
		}

	})

	fillCounter(&resultMS)
	t.Run("Checking the filling of metrics PollCount", func(t *testing.T) {

		if resultMS.PollCount.Type() != "main.counter" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Counter")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if resultMS.PollCount == 0 {
			t.Errorf("The metric %s a value of %v", "PollCount", 0)
		}

	})

	t.Run("Increasing the metric PollCount", func(t *testing.T) {
		var res counter = 1

		if resultMS.PollCount != res {
			t.Errorf("The metric %s has not increased by %v", "PollCount", res)
		}

	})

}

package main

import (
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"strings"
	"testing"
)

func TestFuncServer(t *testing.T) {

	var postStr = "http://127.0.0.1:8080/update/gauge/Alloc/0.1\nhttp://127.0.0.1:8080/update/gauge/" +
		"BuckHashSys/0.002\nhttp://127.0.0.1:8080/update/counter/PollCount/5"

	repository.RefTypeMepStruc()
	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		messageRaz := strings.Split(postStr, "\n")
		if len(messageRaz) != 3 {
			t.Errorf("The string (%s) was incorrectly decomposed into an array", postStr)
		}
	})

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {
		t.Run("Checking the type of the first line", func(t *testing.T) {
			var typeGauge = "gauge"

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]

			if strings.Contains(valElArr, typeGauge) == false {
				t.Errorf("The Gauge type was incorrectly determined")
			}
		})

		t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]

			typeMetric := valStrMetrics(valElArr, 4)
			nameMetric := valStrMetrics(valElArr, 5)
			valueMetric := valStrMetrics(valElArr, 6)

			err := repository.SetValue(typeMetric, nameMetric, valueMetric)
			if err != nil {
				t.Errorf("Error setting the value %s metric %s", valueMetric, nameMetric)
			}

			if repository.Metrics[nameMetric].(repository.Gauge) != repository.Gauge(0.1) {
				t.Errorf("Incorrect definition of the metric %s value %v", "Alloc", valElArr)
			}
		})
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

		t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[2]

			typeMetric := valStrMetrics(valElArr, 4)
			nameMetric := valStrMetrics(valElArr, 5)
			valueMetric := valStrMetrics(valElArr, 6)

			err := repository.SetValue(typeMetric, nameMetric, valueMetric)
			if err != nil {
				t.Errorf("Error setting the value %s metric %s", valueMetric, nameMetric)
			}
			//valueCounter := repository.GetValue(typeMetric, nameMetric)

			if repository.Metrics[nameMetric].(repository.Counter) != repository.Counter(5) {
				t.Errorf("Incorrect definition of the metric %s value %v", "PollCount", valElArr)
			}
		})
	})

}

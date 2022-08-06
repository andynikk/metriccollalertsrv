package main

import (
	"strings"
	"testing"
)

func TestFuncServer(t *testing.T) {

	var postStr = "http://127.0.0.1:8080/update/main.gauge/Alloc/0.1\nhttp://127.0.0.1:8080/update/main.gauge/" +
		"BuckHashSys/0.002\nhttp://127.0.0.1:8080/update/main.counter/PollCount/5"

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		messageRaz := strings.Split(postStr, "\n")
		if len(messageRaz) != 3 {
			t.Errorf("The string (%s) was incorrectly decomposed into an array", postStr)
		}
	})

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {
		t.Run("Checking the type of the first line", func(t *testing.T) {
			var typeGauge = servMetrixStats.Alloc.Type()

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]

			if strings.Contains(valElArr, typeGauge) == false {
				t.Errorf("The Gauge type was incorrectly determined")
			}
		})

		t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]
			valueGauge := valueGauge(valElArr)
			if valueGauge != 0.1 {
				t.Errorf("Incorrect definition of the metric %s value %v", "Alloc", valElArr)
			}
		})
	})

	t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
		t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
			var typeCounter = servMetrixStats.PollCount.Type()

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[2]

			if strings.Contains(valElArr, typeCounter) == false {
				t.Errorf("The Counter type was incorrectly determined")
			}
		})

		t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[2]
			valueCounter := valueGauge(valElArr)
			if valueCounter != 5 {
				t.Errorf("Incorrect definition of the metric %s value %v", "PollCount", valElArr)
			}
		})
	})

}

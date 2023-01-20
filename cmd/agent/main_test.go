package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/andynikk/metriccollalertsrv/internal/agent"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

func TestmakeMsg(adresServer string, memStats agent.MetricsGauge) string {

	const msgFormat = "http://%s/update/%s/%s/%v"

	var msg []string

	val := memStats["Alloc"]
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, val.Type(), "Alloc", 0.1))

	val = memStats["BuckHashSys"]
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, val.Type(), "BuckHashSys", 0.002))

	return strings.Join(msg, "\n")
}

func TestFuncAgen(t *testing.T) {
	config := environment.AgentConfig{}
	config.InitConfigAgentENV()
	config.InitConfigAgentFile()
	config.InitConfigAgentDefault()

	a := agent.GeneralAgent{
		Config:       &config,
		PollCount:    0,
		MetricsGauge: make(agent.MetricsGauge),
	}

	var argErr = "err"

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	t.Run("Checking init config", func(t *testing.T) {
		a.Config = environment.InitConfigAgent()
		if a.Config.Address == "" {
			t.Errorf("Error checking init config")
		}
	})

	t.Run("Checking the structure creation", func(t *testing.T) {

		var realResult agent.MetricsGauge

		if a.MetricsGauge["Alloc"] != realResult["Alloc"] &&
			a.MetricsGauge["RandomValue"] != realResult["RandomValue"] {

			//t.Errorf("Structure creation error", resultMS, realResult)
			t.Errorf("Structure creation error (%s)", argErr)
		}
		t.Run("Creating a submission line", func(t *testing.T) {
			adressServer := a.Config.Address
			var resultStr = fmt.Sprintf("http://%s/update/gauge/Alloc/0.1"+
				"\nhttp://%s/update/gauge/BuckHashSys/0.002", adressServer, adressServer)

			resultMassage := TestmakeMsg(adressServer, realResult)
			if resultStr != resultMassage {
				t.Errorf("Error creating a submission line (%s)", argErr)
			}
		})
	})

	t.Run("Checking rsa crypt", func(t *testing.T) {
		t.Run("Checking init crypto key", func(t *testing.T) {
			a.KeyEncryption, _ = encryption.InitPublicKey(a.Config.CryptoKey)
			if a.Config.CryptoKey != "" && a.KeyEncryption.PublicKey == nil {
				t.Errorf("Error checking init crypto key")
			}
			t.Run("Checking rsa encrypt", func(t *testing.T) {
				testMsg := "Тестовое сообщение"
				_, err := a.KeyEncryption.RsaEncrypt([]byte(testMsg))
				if err != nil {
					t.Errorf("Error checking rsa encrypt")
				}
			})
		})
	})

	t.Run("Checking the filling of metrics", func(t *testing.T) {
		a.FillMetric()
		if len(a.MetricsGauge) == 0 || a.PollCount == 0 {
			t.Errorf("Error checking the filling of metrics")
		}
		t.Run("Checking the filling of other metrics", func(t *testing.T) {
			a.MetricsOtherScan()
			if _, ok := a.MetricsGauge["TotalMemory"]; !ok {
				t.Errorf("Error checking the filling of other metrics")
			}
		})
	})

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		val := a.MetricsGauge["Frees"]
		if val.Type() != "gauge" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Gauge")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if a.MetricsGauge["Alloc"] == 0 {
			t.Errorf("The metric %s a value of %v", "Alloc", 0)
		}

	})

	t.Run("Checking fillings the metrics", func(t *testing.T) {
		mapMetricsButch, err := a.SendMetricsServer()
		if err != nil {
			t.Errorf("Error checking fillings the metrics")
		}
		t.Run("Send metrics to server", func(t *testing.T) {
			for _, allMetrics := range mapMetricsButch {

				gziparrMetrics, err := allMetrics.PrepareMetrics(a.KeyEncryption)
				if err != nil {
					constants.Logger.ErrorLog(err)
					t.Errorf("Send metrics to server")
				}

				resp := httptest.NewRecorder()
				req, err := http.NewRequest("POST", fmt.Sprintf("%s/updates", a.Config.Address),
					strings.NewReader(string(gziparrMetrics)))
				if err != nil {
					t.Fatal(err)
				}
				http.DefaultServeMux.ServeHTTP(resp, req)
				if p, err := io.ReadAll(resp.Body); err != nil {
					t.Errorf("Error send metrics to server")
				} else {
					if string(p) != "" {
						t.Errorf("Error send metrics to server")
					}
				}
			}
		})
	})

	t.Run("Checking the filling of metrics PollCount", func(t *testing.T) {

		val := repository.Counter(a.PollCount)
		if val.Type() != "counter" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Counter")
		}
	})

	t.Run("Checking the metrics value PollCount", func(t *testing.T) {
		if a.PollCount == 0 {
			t.Errorf("The metric %s a value of %v", "PollCount", 0)
		}

	})

	t.Run("Increasing the metric PollCount", func(t *testing.T) {
		var res = int64(1)
		if a.PollCount != res {
			t.Errorf("The metric %s has not increased by %v", "PollCount", res)
		}

	})

}

type Sender interface {
	GoPost2Server(mapMetricsButch agent.MapMetricsButch)
}

func BenchmarkSendMetrics(b *testing.B) {

	config := environment.InitConfigAgent()
	a := agent.NewAgent(config)

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		var allMetrics agent.EmptyArrMetrics
		mapMetricsButch := agent.MapMetricsButch{}

		val := repository.Gauge(0)
		for j := 0; j < 10; j++ {
			val = val + 1
			id := fmt.Sprintf("Metric %d", j)
			floatJ := float64(j)
			metrica := encoding.Metrics{ID: id, MType: val.Type(), Value: &floatJ, Hash: ""}
			allMetrics = append(allMetrics, metrica)
		}
		mapMetricsButch[1] = allMetrics
		wg.Add(1)
		go func() {
			defer wg.Done()
			if config.StringTypeServer == constants.TypeSrvGRPC.String() {
				a.(*agent.AgentGRPC).Post2Server(mapMetricsButch)
			} else {
				a.(*agent.HTTPAgent).Post2Server(mapMetricsButch)
			}
		}()
	}
	wg.Wait()
}

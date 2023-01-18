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

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestmakeMsg(adresServer string, memStats MetricsGauge) string {

	const msgFormat = "http://%s/update/%s/%s/%v"

	var msg []string

	val := memStats["Alloc"]
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, val.Type(), "Alloc", 0.1))

	val = memStats["BuckHashSys"]
	msg = append(msg, fmt.Sprintf(msgFormat, adresServer, val.Type(), "BuckHashSys", 0.002))

	return strings.Join(msg, "\n")
}

func TestFuncAgen(t *testing.T) {
	a := agent{}
	a.data.metricsGauge = make(MetricsGauge)

	var argErr = "err"

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	t.Run("Checking init config", func(t *testing.T) {
		a.config = environment.InitConfigAgent()
		if a.config.Address == "" {
			t.Errorf("Error checking init config")
		}
	})

	t.Run("Checking the structure creation", func(t *testing.T) {

		var realResult MetricsGauge

		if a.data.metricsGauge["Alloc"] != realResult["Alloc"] &&
			a.data.metricsGauge["RandomValue"] != realResult["RandomValue"] {

			//t.Errorf("Structure creation error", resultMS, realResult)
			t.Errorf("Structure creation error (%s)", argErr)
		}
		t.Run("Creating a submission line", func(t *testing.T) {
			adressServer := a.config.Address
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
			a.KeyEncryption, _ = encryption.InitPublicKey(a.config.CryptoKey)
			if a.config.CryptoKey != "" && a.KeyEncryption.PublicKey == nil {
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
		a.fillMetric()
		if len(a.metricsGauge) == 0 || a.pollCount == 0 {
			t.Errorf("Error checking the filling of metrics")
		}
		t.Run("Checking the filling of other metrics", func(t *testing.T) {
			a.metrixOtherScan()
			if _, ok := a.metricsGauge["TotalMemory"]; !ok {
				t.Errorf("Error checking the filling of other metrics")
			}
		})
	})

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		val := a.data.metricsGauge["Frees"]
		if val.Type() != "gauge" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Gauge")
		}
	})

	t.Run("Checking the metrics value Gauge", func(t *testing.T) {
		if a.data.metricsGauge["Alloc"] == 0 {
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

				gziparrMetrics, err := allMetrics.prepareMetrics(a.KeyEncryption)
				if err != nil {
					constants.Logger.ErrorLog(err)
					t.Errorf("Send metrics to server")
				}

				resp := httptest.NewRecorder()
				req, err := http.NewRequest("POST", fmt.Sprintf("%s/updates", a.config.Address),
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

		val := repository.Counter(a.data.pollCount)
		if val.Type() != "counter" {
			t.Errorf("Metric %s is not a type %s", "Frees", "Counter")
		}
	})

	t.Run("Checking the metrics value PollCount", func(t *testing.T) {
		if a.data.pollCount == 0 {
			t.Errorf("The metric %s a value of %v", "PollCount", 0)
		}

	})

	t.Run("Increasing the metric PollCount", func(t *testing.T) {
		var res = int64(1)
		if a.data.pollCount != res {
			t.Errorf("The metric %s has not increased by %v", "PollCount", res)
		}

	})

}

func BenchmarkSendMetrics(b *testing.B) {
	a := agent{}
	a.config = environment.InitConfigAgent()
	if a.config.Address == "" {
		return
	}

	if a.config.StringTypeServer == constants.TypeSrvGRPC.String() {
		conn, err := grpc.Dial(constants.AddressServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		a.GRPCClientConn = conn
	}

	certPublicKey, _ := encryption.InitPublicKey(a.config.CryptoKey)
	a.KeyEncryption = certPublicKey

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		var allMetrics emptyArrMetrics
		mapMetricsButch := MapMetricsButch{}

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
			a.goPost2Server(mapMetricsButch)
		}()
	}
	wg.Wait()
}

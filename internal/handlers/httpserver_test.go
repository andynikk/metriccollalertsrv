package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

var srv Server

func ExampleRepStore_HandlerGetAllMetrics() {
	r := srv.GetRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/", strings.NewReader(""))
	if err != nil {
		return
	}
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	msg := fmt.Sprintf("Metrics: %s. HTTP-Status: %d",
		resp.Header.Get("Metrics-Val"), resp.StatusCode)
	fmt.Println(msg)

	// Output:
	// Metrics: TestGauge = 0.001. HTTP-Status: 200
}

func ExampleRepStore_HandlerSetMetricaPOST() {

	ts := httptest.NewServer(srv.GetRouter())
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL+"/update/gauge/TestGauge/0.01", strings.NewReader(""))
	if err != nil {
		return
	}
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fmt.Print(resp.StatusCode)

	// Output:
	// 200
}

func ExampleRepStore_HandlerGetValue() {

	ts := httptest.NewServer(srv.GetRouter())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/value/gauge/TestGauge", strings.NewReader(""))
	if err != nil {
		return
	}
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fmt.Print(resp.StatusCode)

	// Output:
	// 200
}

func init() {

	config := environment.ServerConfig{}
	config.InitConfigServerENV()
	config.InitConfigServerFile()
	config.InitConfigServerDefault()

	config.StorageType, _ = repository.InitStoreDB(config.StorageType, config.DatabaseDsn)
	config.StorageType, _ = repository.InitStoreFile(config.StorageType, config.StoreFile)

	srv = NewServer(&config)
	rp := srv.GetRepStore()

	valG := repository.Gauge(0)
	if ok := valG.SetFromText("0.001"); !ok {
		return
	}
	rp.MutexRepo["TestGauge"] = &valG
}

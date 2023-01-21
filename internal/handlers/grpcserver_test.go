package handlers

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestFuncServer(t *testing.T) {

	config := environment.InitConfigServer()
	server := newGRPCServer(config)

	t.Run("Checking init server", func(t *testing.T) {
		if config.Address == "" {
			t.Errorf("Error checking init server")
		}
	})

	var IPAddress string
	t.Run("Checking get current IP", func(t *testing.T) {
		hn, _ := os.Hostname()
		IPs, _ := net.LookupIP(hn)
		IPAddress = networks.IPv4RangesToStr(IPs)
		if IPAddress == "" {
			t.Errorf("Error checking get current IP")
		}
	})

	mHeader := map[string]string{"Content-Type": "application/json",
		"Content-Encoding": "gzip"}
	if server.PK != nil && server.PK.PrivateKey != nil && server.PK.PublicKey != nil {
		mHeader["Content-Encryption"] = server.PK.TypeEncryption
	}

	md := metadata.New(mHeader)
	ctx := context.TODO()

	ctx = metadata.NewOutgoingContext(ctx, md)

	t.Run("Checking handlers PING", func(t *testing.T) {
		req := EmptyRequest{}
		textErr, err := server.PingDataBase(ctx, &req)
		if errs.CodeGRPC(err) != codes.OK && server.Config.Storage == nil {
			t.Errorf("Error checking handlers PING. %s", textErr)
		}
	})

	t.Run("Checking handlers Update", func(t *testing.T) {
		tests := []struct {
			name           string
			request        *ResponseProperties
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения counter", request: &ResponseProperties{MetType: []byte("counter"),
				MetName: []byte("testSetGet332"), MetValue: []byte("6")}, wantStatusCode: codes.OK},
			{name: "Проверка на не правильный тип метрики", request: &ResponseProperties{MetType: []byte("notcounter"),
				MetName: []byte("testSetGet332"), MetValue: []byte("6")}, wantStatusCode: codes.Unimplemented},
			{name: "Проверка на не правильное значение метрики", request: &ResponseProperties{MetType: []byte("counter"),
				MetName: []byte("testSetGet332"), MetValue: []byte("non")}, wantStatusCode: codes.PermissionDenied},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				textErr, err := server.UpdateOneMetrics(ctx, tt.request)
				if errs.CodeGRPC(err) != tt.wantStatusCode {
					t.Errorf("Error checking handlers Update (%s). %s", textErr, tt.name)
				}
			})
		}
	})

	t.Run("Checking handlers Update JSON", func(t *testing.T) {
		tests := []struct {
			name           string
			request        encoding.Metrics
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения gauge", request: testMetricsGouge(server.Config.Key),
				wantStatusCode: codes.OK},
			{name: "Проверка на установку значения counter", request: testMetricsCaunter(server.Config.Key),
				wantStatusCode: codes.OK},
			{name: "Проверка на не правильный тип метрики gauge", request: testMetricsNoGouge(server.Config.Key),
				wantStatusCode: codes.Unimplemented},
			{name: "Проверка на не правильный тип метрики counter", request: testMetricsNoCounter(server.Config.Key),
				wantStatusCode: codes.Unimplemented},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var gziparrMetrics []byte
				//var storedData encoding.ArrMetrics
				//storedData = append(storedData, tt.request)

				t.Run("Checking gzip", func(t *testing.T) {
					arrMetrics, err := json.MarshalIndent(tt.request, "", " ")
					if err != nil {
						t.Errorf("Error checking gzip. %s", tt.name)
					}

					gziparrMetrics, err = compression.Compress(arrMetrics)
					if err != nil {
						t.Errorf("Error checking gzip. %s", tt.name)
					}

				})

				req := RequestUpdateByte{Body: gziparrMetrics}
				key := KeyContext("content-encoding")
				ctxValue := context.WithValue(ctx, key, "gzip")
				_, err := server.UpdateOneMetricsJSON(ctxValue, &req)
				if errs.CodeGRPC(err) != tt.wantStatusCode {
					t.Errorf("Error checking handlers Update JSON. %s", tt.name)
				}
			})
		}
	})

	t.Run("Checking handlers Updates JSON", func(t *testing.T) {
		var storedData encoding.ArrMetrics
		storedData = append(storedData, testMetricsGouge(server.Config.Key))
		storedData = append(storedData, testMetricsCaunter(server.Config.Key))

		arrMetrics, err := json.MarshalIndent(storedData, "", " ")
		if err != nil {
			t.Errorf("Error checking gzip. %s", "Updates JSON")
		}
		gziparrMetrics, err := compression.Compress(arrMetrics)
		if err != nil {
			t.Errorf("Error checking gzip. %s", "Updates JSON")
		}

		req := RequestUpdateByte{Body: gziparrMetrics}
		key := KeyContext("content-encoding")
		ctxValue := context.WithValue(ctx, key, "gzip")
		_, err = server.UpdatesAllMetricsJSON(ctxValue, &req)
		if errs.CodeGRPC(err) != codes.OK {
			t.Errorf("Error checking handlers Update JSON.")
		}
	})

	t.Run("Checking handlers Value JSON", func(t *testing.T) {

		tests := []struct {
			name           string
			request        encoding.Metrics
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения gauge", request: testMetricsGouge(server.Config.Key),
				wantStatusCode: codes.OK},
			{name: "Проверка на установку значения counter", request: testMetricsCaunter(server.Config.Key),
				wantStatusCode: codes.OK},
			{name: "Проверка на не правильное значение метрики gauge", request: testMetricsWrongGouge(server.Config.Key),
				wantStatusCode: codes.NotFound},
			{name: "Проверка на не правильное значение метрики counter", request: testMetricsWrongCounter(server.Config.Key),
				wantStatusCode: codes.NotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				arrMetrics, err := json.MarshalIndent(tt.request, "", " ")
				if err != nil {
					t.Errorf("Error checking gzip. %s", tt.name)
				}

				gziparrMetrics, err := compression.Compress(arrMetrics)
				if err != nil {
					t.Errorf("Error checking gzip. %s", tt.name)
				}

				req := RequestByte{Body: gziparrMetrics}
				key := KeyContext("content-encoding")
				ctxValue := context.WithValue(ctx, key, "gzip")
				_, err = server.GetValueJSON(ctxValue, &req)
				if errs.CodeGRPC(err) != tt.wantStatusCode {
					t.Errorf("Error checking handlers Value JSON. %s", tt.name)
				}
			})
		}
	})

	t.Run("Checking handlers Value", func(t *testing.T) {

		tests := []struct {
			name           string
			request        string
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения gauge", request: testMetricsGouge(server.Config.Key).ID,
				wantStatusCode: codes.OK},
			{name: "Проверка на установку значения counter", request: testMetricsCaunter(server.Config.Key).ID,
				wantStatusCode: codes.OK},
			{name: "Проверка на не правильное значение метрики gauge", request: testMetricsWrongGouge(server.Config.Key).ID,
				wantStatusCode: codes.NotFound},
			{name: "Проверка на не правильное значение метрики counter", request: testMetricsWrongCounter(server.Config.Key).ID,
				wantStatusCode: codes.NotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				req := ResponseProperties{MetName: []byte(tt.request)}
				rep, err := server.GetValue(ctx, &req)
				if errs.CodeGRPC(err) != tt.wantStatusCode {
					t.Errorf("Error checking handlers Value (%s). %s", rep.Result, tt.name)
				}
			})
		}
	})

	t.Run("Checking handlers ListMetrics", func(t *testing.T) {

		req := EmptyRequest{}
		res, err := server.GetListMetrics(ctx, &req)

		if err != nil || res == nil {
			t.Errorf("Error checking handlers ListMetrics.")
		}
	})

}

func testMetricsGouge(configKey string) encoding.Metrics {

	var fValue = 0.001

	var mGauge encoding.Metrics
	mGauge.ID = "TestGauge"
	mGauge.MType = "gauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.ID, configKey)

	return mGauge
}

func testMetricsWrongGouge(configKey string) encoding.Metrics {

	var fValue = 0.001

	var mGauge encoding.Metrics
	mGauge.ID = "TestGauge322"
	mGauge.MType = "gauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.ID, configKey)

	return mGauge
}

func testMetricsNoGouge(configKey string) encoding.Metrics {

	var fValue = 0.001

	var mGauge encoding.Metrics
	mGauge.ID = "TestGauge"
	mGauge.MType = "nogauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.ID, configKey)

	return mGauge
}

func testMetricsCaunter(configKey string) encoding.Metrics {
	var iDelta int64 = 10

	var mCounter encoding.Metrics
	mCounter.ID = "TestCounter"
	mCounter.MType = "counter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.ID, configKey)

	return mCounter
}

func testMetricsNoCounter(configKey string) encoding.Metrics {
	var iDelta int64 = 10

	var mCounter encoding.Metrics
	mCounter.ID = "TestCounter"
	mCounter.MType = "nocounter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.ID, configKey)

	return mCounter
}

func testMetricsWrongCounter(configKey string) encoding.Metrics {
	var iDelta int64 = 10

	var mCounter encoding.Metrics
	mCounter.ID = "TestCounter322"
	mCounter.MType = "counter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.ID, configKey)

	return mCounter
}

package handlers

import (
	"context"
	"testing"

	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type KeyContext string

func TestFuncServer(t *testing.T) {

	config := environment.InitConfigServer()
	server := newGRPCServer(config)

	t.Run("Checking init server", func(t *testing.T) {
		if config.Address == "" {
			t.Errorf("Error checking init server")
		}
	})

	t.Run("Checking get current IP", func(t *testing.T) {
		IPAddress, err := environment.GetLocalIPAddress(config.Address)
		if err != nil && IPAddress == "" {
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
		req := &emptypb.Empty{}
		textErr, err := server.PingDataBase(ctx, req)
		if errs.CodeGRPC(err) != codes.OK && server.Config.Storage == nil {
			t.Errorf("Error checking handlers PING. %s", textErr)
		}
	})

	t.Run("Checking handlers Update", func(t *testing.T) {
		tests := []struct {
			name           string
			request        *pb.RequestMetricsString
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения counter", request: &pb.RequestMetricsString{MetricsType: "counter",
				MetricsName: "testSetGet332", MetricsValue: "6"}, wantStatusCode: codes.OK},
			{name: "Проверка на не правильный тип метрики", request: &pb.RequestMetricsString{MetricsType: "notcounter",
				MetricsName: "testSetGet332", MetricsValue: "6"}, wantStatusCode: codes.Unimplemented},
			{name: "Проверка на не правильное значение метрики", request: &pb.RequestMetricsString{MetricsType: "counter",
				MetricsName: "testSetGet332", MetricsValue: "non"}, wantStatusCode: codes.PermissionDenied},
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
			request        *pb.Metrics
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
				req := pb.RequestMetrics{Metrics: tt.request}
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
		var storedData []*pb.Metrics
		storedData = append(storedData, testMetricsGouge(server.Config.Key))
		storedData = append(storedData, testMetricsCaunter(server.Config.Key))

		req := pb.RequestListMetrics{Metrics: storedData}
		key := KeyContext("content-encoding")
		ctxValue := context.WithValue(ctx, key, "gzip")
		_, err := server.UpdatesAllMetricsJSON(ctxValue, &req)
		if errs.CodeGRPC(err) != codes.OK {
			t.Errorf("Error checking handlers Update JSON.")
		}
	})

	t.Run("Checking handlers Value JSON", func(t *testing.T) {

		tests := []struct {
			name           string
			request        *pb.GetMetrics
			wantStatusCode codes.Code
		}{
			{name: "Проверка на установку значения gauge", request: testGetMetricsGouge(),
				wantStatusCode: codes.OK},
			{name: "Проверка на установку значения counter", request: testGetMetricsCaunter(),
				wantStatusCode: codes.OK},
			{name: "Проверка на не правильное значение метрики gauge", request: testGetMetricsWrongGouge(),
				wantStatusCode: codes.NotFound},
			{name: "Проверка на не правильное значение метрики counter", request: testGetMetricsWrongCounter(),
				wantStatusCode: codes.NotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				req := pb.RequestGetMetrics{Metrics: tt.request}
				key := KeyContext("content-encoding")
				ctxValue := context.WithValue(ctx, key, "gzip")
				_, err := server.GetValueJSON(ctxValue, &req)
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
			{name: "Проверка на установку значения gauge", request: testMetricsGouge(server.Config.Key).Id,
				wantStatusCode: codes.OK},
			{name: "Проверка на установку значения counter", request: testMetricsCaunter(server.Config.Key).Id,
				wantStatusCode: codes.OK},
			{name: "Проверка на не правильное значение метрики gauge", request: testMetricsWrongGouge(server.Config.Key).Id,
				wantStatusCode: codes.NotFound},
			{name: "Проверка на не правильное значение метрики counter", request: testMetricsWrongCounter(server.Config.Key).Id,
				wantStatusCode: codes.NotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				req := pb.RequestMetricsName{MetricsName: tt.request}
				rep, err := server.GetValue(ctx, &req)
				if errs.CodeGRPC(err) != tt.wantStatusCode {
					t.Errorf("Error checking handlers Value (%s). %s", rep.Result, tt.name)
				}
			})
		}
	})

	t.Run("Checking handlers ListMetrics", func(t *testing.T) {

		req := emptypb.Empty{}
		res, err := server.GetListMetrics(ctx, &req)

		if err != nil || len(res.Metrics) == 0 {
			t.Errorf("Error checking handlers ListMetrics.")
		}
	})

}

func testGetMetricsGouge() *pb.GetMetrics {

	var mGauge pb.GetMetrics
	mGauge.Id = "TestGauge"
	mGauge.Mtype = pb.GetMetrics_MType(0)

	return &mGauge
}

func testMetricsGouge(configKey string) *pb.Metrics {

	var fValue = 0.001

	var mGauge pb.Metrics
	mGauge.Id = "TestGauge"
	mGauge.Mtype = "gauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.Id, configKey)

	return &mGauge
}

func testGetMetricsWrongGouge() *pb.GetMetrics {

	var mGauge pb.GetMetrics
	mGauge.Id = "TestGauge322"
	mGauge.Mtype = pb.GetMetrics_MType(0)

	return &mGauge
}

func testMetricsWrongGouge(configKey string) *pb.Metrics {

	var fValue = 0.001

	var mGauge pb.Metrics
	mGauge.Id = "TestGauge322"
	mGauge.Mtype = "gauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.Id, configKey)

	return &mGauge
}

func testMetricsNoGouge(configKey string) *pb.Metrics {

	var fValue = 0.001

	var mGauge pb.Metrics
	mGauge.Id = "TestGauge"
	mGauge.Mtype = "nogauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.Id, configKey)

	return &mGauge
}

func testGetMetricsCaunter() *pb.GetMetrics {

	var mCounter pb.GetMetrics
	mCounter.Id = "TestCounter"
	mCounter.Mtype = pb.GetMetrics_MType(2)

	return &mCounter
}

func testMetricsCaunter(configKey string) *pb.Metrics {
	var iDelta int64 = 10

	var mCounter pb.Metrics
	mCounter.Id = "TestCounter"
	mCounter.Mtype = "counter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.Id, configKey)

	return &mCounter
}

func testMetricsNoCounter(configKey string) *pb.Metrics {
	var iDelta int64 = 10

	var mCounter pb.Metrics
	mCounter.Id = "TestCounter"
	mCounter.Mtype = "nocounter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.Id, configKey)

	return &mCounter
}

func testGetMetricsWrongCounter() *pb.GetMetrics {

	var mCounter pb.GetMetrics
	mCounter.Id = "TestCounter322"
	mCounter.Mtype = pb.GetMetrics_MType(2)

	return &mCounter
}

func testMetricsWrongCounter(configKey string) *pb.Metrics {
	var iDelta int64 = 10

	var mCounter pb.Metrics
	mCounter.Id = "TestCounter322"
	mCounter.Mtype = "counter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.Id, configKey)

	return &mCounter
}

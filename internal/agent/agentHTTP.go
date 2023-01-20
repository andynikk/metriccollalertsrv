package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
)

type HTTPAgent struct {
	GeneralAgent
}

func newAgentHTTP(configAgent *environment.AgentConfig) *HTTPAgent {
	//pk, err := encryption.InitPublicKey(configAgent.CryptoKey)
	//if err != nil {
	//	constants.Logger.ErrorLog(err)
	//	return nil
	//}

	a := HTTPAgent{
		GeneralAgent: GeneralAgent{
			Config:       configAgent,
			PollCount:    0,
			MetricsGauge: make(MetricsGauge),
			//KeyEncryption: pk,
		},
	}
	return &a
}

func (a *HTTPAgent) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go a.GeneralAgent.GoMetricsScan(ctx, cancelFunc)
	go a.GeneralAgent.GoMetricsOtherScan(ctx, cancelFunc)
	go a.GoMakeRequest(ctx, cancelFunc)
}

func (a *HTTPAgent) Stop() {
	mapMetricsButch, _ := a.SendMetricsServer()
	a.Post2Server(mapMetricsButch)
}

func (a *HTTPAgent) Post2Server(metricsButch MapMetricsButch) {

	for _, metrics := range metricsButch {

		gzipArrMetrics, err := metrics.PrepareMetrics(a.KeyEncryption)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		addressPost := fmt.Sprintf("http://%s/updates", a.Config.Address)
		req, err := http.NewRequest("POST", addressPost, bytes.NewReader(gzipArrMetrics))
		if err != nil {

			constants.Logger.ErrorLog(err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("X-Real-IP", a.Config.IPAddress)
		if a.KeyEncryption != nil && a.KeyEncryption.PublicKey != nil {
			req.Header.Set("Content-Encryption", a.KeyEncryption.TypeEncryption)
		}
		defer req.Body.Close()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
		defer resp.Body.Close()
	}
}

func (a *HTTPAgent) GoMakeRequest(ctx context.Context, cancelFunc context.CancelFunc) {

	reportTicker := time.NewTicker(a.Config.ReportInterval)

	for {
		select {
		case <-reportTicker.C:
			mapAllMetrics, _ := a.SendMetricsServer()
			go a.Post2Server(mapAllMetrics)

		case <-ctx.Done():

			cancelFunc()
			return

		}
	}
}

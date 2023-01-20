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

type AgentHTTP struct {
	GeneralAgent
}

func newAgentHTTP(configAgent *environment.AgentConfig) *AgentHTTP {
	//pk, err := encryption.InitPublicKey(configAgent.CryptoKey)
	//if err != nil {
	//	constants.Logger.ErrorLog(err)
	//	return nil
	//}

	a := AgentHTTP{
		GeneralAgent: GeneralAgent{
			Config:       configAgent,
			PollCount:    0,
			MetricsGauge: make(MetricsGauge),
			//KeyEncryption: pk,
		},
	}

	return &a
}

func (a *AgentHTTP) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go a.GeneralAgent.GoMetricsScan(ctx, cancelFunc)
	go a.GeneralAgent.GoMetricsOtherScan(ctx, cancelFunc)
	go a.GoMakeRequest(ctx, cancelFunc)
}

func (a *AgentHTTP) Stop() {
	mapMetricsButch, _ := a.SendMetricsServer()
	a.Post2Server(mapMetricsButch)
}

func (a *AgentHTTP) Post2Server(metricsButch MapMetricsButch) {

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

func (a *AgentHTTP) GoMakeRequest(ctx context.Context, cancelFunc context.CancelFunc) {

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

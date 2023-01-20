package agent

import (
	"context"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AgentGRPC struct {
	GeneralAgent
	GRPCClientConn *grpc.ClientConn
}

func newAgentGRPC(configAgent *environment.AgentConfig) *AgentGRPC {
	conn, err := grpc.Dial(constants.AddressServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil
	}
	pk, err := encryption.InitPublicKey(configAgent.CryptoKey)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil
	}

	a := AgentGRPC{
		GeneralAgent: GeneralAgent{
			Config:        configAgent,
			PollCount:     0,
			MetricsGauge:  make(MetricsGauge),
			KeyEncryption: pk,
		},
		GRPCClientConn: conn,
	}

	return &a
}

func (a *AgentGRPC) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go a.GeneralAgent.GoMetricsScan(ctx, cancelFunc)
	go a.GeneralAgent.GoMetricsOtherScan(ctx, cancelFunc)
	go a.GoMakeRequest(ctx, cancelFunc)
}

func (a *AgentGRPC) Stop() {
	mapMetricsButch, _ := a.SendMetricsServer()
	a.Post2Server(mapMetricsButch)
}

func (a *AgentGRPC) Post2Server(metricsButch MapMetricsButch) {

	for _, metrics := range metricsButch {

		gzipArrMetrics, err := metrics.PrepareMetrics(a.KeyEncryption)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		c := handlers.NewMetricCollectorClient(a.GRPCClientConn)
		mHeader := map[string]string{"Content-Type": "application/json",
			"Content-Encoding": "gzip",
			"X-Real-IP":        a.Config.IPAddress}
		if a.KeyEncryption != nil && a.KeyEncryption.PublicKey != nil {
			mHeader["Content-Encryption"] = a.KeyEncryption.TypeEncryption
		}

		md := metadata.New(mHeader)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		_, err = c.UpdatesAllMetricsJSON(ctx, &handlers.UpdatesRequest{Body: gzipArrMetrics})
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
	}
}

func (a *AgentGRPC) GoMakeRequest(ctx context.Context, cancelFunc context.CancelFunc) {

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

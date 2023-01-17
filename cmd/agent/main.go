package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricsGauge map[string]repository.Gauge
type emptyArrMetrics []encoding.Metrics
type MapMetricsButch map[int]emptyArrMetrics

type data struct {
	sync.RWMutex
	pollCount    int64
	metricsGauge MetricsGauge
}

type ServerHTTP bool
type ServerGRPC bool

type ServerType interface {
	Post2Server(agent *agent, allMetrics []byte) error
}

type agent struct {
	config        *environment.AgentConfig
	KeyEncryption *encryption.KeyEncryption
	data
}

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func (eam *emptyArrMetrics) prepareMetrics(key *encryption.KeyEncryption) ([]byte, error) {
	arrMetrics, err := json.MarshalIndent(eam, "", " ")
	if err != nil {
		return nil, err
	}

	gziparrMetrics, err := compression.Compress(arrMetrics)
	if err != nil {
		return nil, err
	}

	if key != nil && key.PublicKey != nil {
		gziparrMetrics, err = key.RsaEncrypt(gziparrMetrics)
		if err != nil {
			return nil, err
		}
	}

	return gziparrMetrics, nil
}

func (a *agent) fillMetric() {

	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)

	a.data.Lock()
	defer a.data.Unlock()

	a.data.metricsGauge["Alloc"] = repository.Gauge(mems.Alloc)
	a.data.metricsGauge["BuckHashSys"] = repository.Gauge(mems.BuckHashSys)
	a.data.metricsGauge["Frees"] = repository.Gauge(mems.Frees)
	a.data.metricsGauge["GCCPUFraction"] = repository.Gauge(mems.GCCPUFraction)
	a.data.metricsGauge["GCSys"] = repository.Gauge(mems.GCSys)
	a.data.metricsGauge["HeapAlloc"] = repository.Gauge(mems.HeapAlloc)
	a.data.metricsGauge["HeapIdle"] = repository.Gauge(mems.HeapIdle)
	a.data.metricsGauge["HeapInuse"] = repository.Gauge(mems.HeapInuse)
	a.data.metricsGauge["HeapObjects"] = repository.Gauge(mems.HeapObjects)
	a.data.metricsGauge["HeapReleased"] = repository.Gauge(mems.HeapReleased)
	a.data.metricsGauge["HeapSys"] = repository.Gauge(mems.HeapSys)
	a.data.metricsGauge["LastGC"] = repository.Gauge(mems.LastGC)
	a.data.metricsGauge["Lookups"] = repository.Gauge(mems.Lookups)
	a.data.metricsGauge["MCacheInuse"] = repository.Gauge(mems.MCacheInuse)
	a.data.metricsGauge["MCacheSys"] = repository.Gauge(mems.MCacheSys)
	a.data.metricsGauge["MSpanInuse"] = repository.Gauge(mems.MSpanInuse)
	a.data.metricsGauge["MSpanSys"] = repository.Gauge(mems.MSpanSys)
	a.data.metricsGauge["Mallocs"] = repository.Gauge(mems.Mallocs)
	a.data.metricsGauge["NextGC"] = repository.Gauge(mems.NextGC)
	a.data.metricsGauge["NumForcedGC"] = repository.Gauge(mems.NumForcedGC)
	a.data.metricsGauge["NumGC"] = repository.Gauge(mems.NumGC)
	a.data.metricsGauge["OtherSys"] = repository.Gauge(mems.OtherSys)
	a.data.metricsGauge["PauseTotalNs"] = repository.Gauge(mems.PauseTotalNs)
	a.data.metricsGauge["StackInuse"] = repository.Gauge(mems.StackInuse)
	a.data.metricsGauge["StackSys"] = repository.Gauge(mems.StackSys)
	a.data.metricsGauge["Sys"] = repository.Gauge(mems.Sys)
	a.data.metricsGauge["TotalAlloc"] = repository.Gauge(mems.TotalAlloc)
	a.data.metricsGauge["RandomValue"] = repository.Gauge(rand.Float64())

	a.data.pollCount = a.data.pollCount + 1
}

func (a *agent) metrixOtherScan() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cpuUtilization, _ := cpu.Percent(2*time.Second, false)
	swapMemory, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	CPUutilization1 := repository.Gauge(0)
	for _, val := range cpuUtilization {
		CPUutilization1 = repository.Gauge(val)
		break
	}
	a.data.Lock()

	a.data.metricsGauge["TotalMemory"] = repository.Gauge(swapMemory.Total)
	a.data.metricsGauge["FreeMemory"] = repository.Gauge(swapMemory.Free) + repository.Gauge(rand.Float64())
	a.data.metricsGauge["CPUutilization1"] = CPUutilization1

	a.data.Unlock()
}

func (a *agent) goMetrixOtherScan(ctx context.Context, cancelFunc context.CancelFunc) {

	saveTicker := time.NewTicker(a.config.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			a.metrixOtherScan()

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (a *agent) goMetrixScan(ctx context.Context, cancelFunc context.CancelFunc) {

	saveTicker := time.NewTicker(a.config.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			a.fillMetric()

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (a *agent) goPost2Server(metricsButch MapMetricsButch) error {

	for _, allMetrics := range metricsButch {

		gzipArrMetrics, err := allMetrics.prepareMetrics(a.KeyEncryption)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return err
		}

		//typeSrv := constants.TypeServer.String()
		typeSrv := "HTTP"
		if err = GetTypeSrv(typeSrv).Post2Server(a, gzipArrMetrics); err != nil {
			constants.Logger.ErrorLog(err)
			return err
		}
	}
	return nil
}

func (a *agent) SendMetricsServer() (MapMetricsButch, error) {
	a.data.RLock()
	defer a.data.RUnlock()

	mapMatricsButch := MapMetricsButch{}

	allMetrics := make(emptyArrMetrics, 0)
	i := 0
	sch := 0
	tempMetricsGauge := &a.data.metricsGauge
	for key, val := range *tempMetricsGauge {
		valFloat64 := float64(val)

		msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
		heshVal := cryptohash.HashSHA256(msg, a.config.Key)

		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
		allMetrics = append(allMetrics, metrica)

		i++
		if i == constants.ButchSize {

			mapMatricsButch[sch] = allMetrics
			allMetrics = make(emptyArrMetrics, 0)
			sch++
			i = 0
		}
	}

	cPollCount := repository.Counter(a.data.pollCount)
	msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.data.pollCount)
	heshVal := cryptohash.HashSHA256(msg, a.config.Key)

	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.data.pollCount, Hash: heshVal}
	allMetrics = append(allMetrics, metrica)

	mapMatricsButch[sch] = allMetrics

	return mapMatricsButch, nil
}

func (a *agent) goMakeRequest(ctx context.Context, cancelFunc context.CancelFunc) {

	reportTicker := time.NewTicker(a.config.ReportInterval)

	for {
		select {
		case <-reportTicker.C:
			mapAllMetrics, _ := a.SendMetricsServer()
			go a.goPost2Server(mapAllMetrics)

		case <-ctx.Done():

			cancelFunc()
			return

		}
	}
}

func (s *ServerHTTP) Post2Server(agent *agent, allMetrics []byte) error {

	addressPost := fmt.Sprintf("http://%s/updates", agent.config.Address)

	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(allMetrics))
	if err != nil {

		constants.Logger.ErrorLog(err)

		return errors.New("-- ошибка отправки данных на сервер (1)")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Real-IP", agent.config.IPAddress)
	if agent.KeyEncryption != nil && agent.KeyEncryption.PublicKey != nil {
		req.Header.Set("Content-Encryption", agent.KeyEncryption.TypeEncryption)
	}

	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return errors.New("-- ошибка отправки данных на сервер (2)")
	}
	defer resp.Body.Close()

	return nil
}

func GetTypeSrv(stringTypeServer string) ServerType {

	return new(ServerHTTP)
}

func main() {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	configAgent := environment.InitConfigAgent()
	certPublicKey, _ := encryption.InitPublicKey(configAgent.CryptoKey)

	a := agent{
		config: configAgent,
		data: data{
			pollCount:    0,
			metricsGauge: make(MetricsGauge),
		},
		KeyEncryption: certPublicKey,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	go a.goMetrixScan(ctx, cancelFunc)
	go a.goMetrixOtherScan(ctx, cancelFunc)
	go a.goMakeRequest(ctx, cancelFunc)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop

	//a.metricsGauge["BuckHashSys"] = repository.Gauge(0.0001)
	mapMetricsButch, _ := a.SendMetricsServer()
	_ = a.goPost2Server(mapMetricsButch)
}

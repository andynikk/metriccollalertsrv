package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
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
type EmptyArrMetrics []encoding.Metrics
type MapMetricsButch map[int]EmptyArrMetrics

type Agent interface {
	Run()
	Stop()
}

type GeneralAgent struct {
	Config        *environment.AgentConfig
	KeyEncryption *encryption.KeyEncryption
	sync.RWMutex
	PollCount    int64
	MetricsGauge MetricsGauge
}

func (eam *EmptyArrMetrics) PrepareMetrics(key *encryption.KeyEncryption) ([]byte, error) {
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

func (a *GeneralAgent) FillMetric() {

	var mems runtime.MemStats
	runtime.ReadMemStats(&mems)

	a.Lock()
	defer a.Unlock()

	a.MetricsGauge["Alloc"] = repository.Gauge(mems.Alloc)
	a.MetricsGauge["BuckHashSys"] = repository.Gauge(mems.BuckHashSys)
	a.MetricsGauge["Frees"] = repository.Gauge(mems.Frees)
	a.MetricsGauge["GCCPUFraction"] = repository.Gauge(mems.GCCPUFraction)
	a.MetricsGauge["GCSys"] = repository.Gauge(mems.GCSys)
	a.MetricsGauge["HeapAlloc"] = repository.Gauge(mems.HeapAlloc)
	a.MetricsGauge["HeapIdle"] = repository.Gauge(mems.HeapIdle)
	a.MetricsGauge["HeapInuse"] = repository.Gauge(mems.HeapInuse)
	a.MetricsGauge["HeapObjects"] = repository.Gauge(mems.HeapObjects)
	a.MetricsGauge["HeapReleased"] = repository.Gauge(mems.HeapReleased)
	a.MetricsGauge["HeapSys"] = repository.Gauge(mems.HeapSys)
	a.MetricsGauge["LastGC"] = repository.Gauge(mems.LastGC)
	a.MetricsGauge["Lookups"] = repository.Gauge(mems.Lookups)
	a.MetricsGauge["MCacheInuse"] = repository.Gauge(mems.MCacheInuse)
	a.MetricsGauge["MCacheSys"] = repository.Gauge(mems.MCacheSys)
	a.MetricsGauge["MSpanInuse"] = repository.Gauge(mems.MSpanInuse)
	a.MetricsGauge["MSpanSys"] = repository.Gauge(mems.MSpanSys)
	a.MetricsGauge["Mallocs"] = repository.Gauge(mems.Mallocs)
	a.MetricsGauge["NextGC"] = repository.Gauge(mems.NextGC)
	a.MetricsGauge["NumForcedGC"] = repository.Gauge(mems.NumForcedGC)
	a.MetricsGauge["NumGC"] = repository.Gauge(mems.NumGC)
	a.MetricsGauge["OtherSys"] = repository.Gauge(mems.OtherSys)
	a.MetricsGauge["PauseTotalNs"] = repository.Gauge(mems.PauseTotalNs)
	a.MetricsGauge["StackInuse"] = repository.Gauge(mems.StackInuse)
	a.MetricsGauge["StackSys"] = repository.Gauge(mems.StackSys)
	a.MetricsGauge["Sys"] = repository.Gauge(mems.Sys)
	a.MetricsGauge["TotalAlloc"] = repository.Gauge(mems.TotalAlloc)
	a.MetricsGauge["RandomValue"] = repository.Gauge(rand.Float64())

	a.PollCount = a.PollCount + 1
}

func (a *GeneralAgent) MetricsOtherScan() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cpuUtilization, _ := cpu.Percent(2*time.Second, false)
	swapMemory, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	CPUUtilization1 := repository.Gauge(0)
	for _, val := range cpuUtilization {
		CPUUtilization1 = repository.Gauge(val)
		break
	}
	a.Lock()
	defer a.Unlock()

	a.MetricsGauge["TotalMemory"] = repository.Gauge(swapMemory.Total)
	a.MetricsGauge["FreeMemory"] = repository.Gauge(swapMemory.Free) + repository.Gauge(rand.Float64())
	a.MetricsGauge["CPUutilization1"] = CPUUtilization1
}

func (a *GeneralAgent) SendMetricsServer() (MapMetricsButch, error) {
	a.RLock()
	defer a.RUnlock()

	mapMatricsButch := MapMetricsButch{}

	allMetrics := make(EmptyArrMetrics, 0)
	i := 0
	sch := 0
	tempMetricsGauge := &a.MetricsGauge
	for key, val := range *tempMetricsGauge {
		valFloat64 := float64(val)

		msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
		heshVal := cryptohash.HashSHA256(msg, a.Config.Key)

		metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
		allMetrics = append(allMetrics, metrica)

		i++
		if i == constants.ButchSize {

			mapMatricsButch[sch] = allMetrics
			allMetrics = make(EmptyArrMetrics, 0)
			sch++
			i = 0
		}
	}

	cPollCount := repository.Counter(a.PollCount)
	msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.PollCount)
	heshVal := cryptohash.HashSHA256(msg, a.Config.Key)

	metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.PollCount, Hash: heshVal}
	allMetrics = append(allMetrics, metrica)

	mapMatricsButch[sch] = allMetrics

	return mapMatricsButch, nil
}

func (a *GeneralAgent) GoMetricsOtherScan(ctx context.Context, cancelFunc context.CancelFunc) {

	saveTicker := time.NewTicker(a.Config.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			a.MetricsOtherScan()

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (a *GeneralAgent) GoMetricsScan(ctx context.Context, cancelFunc context.CancelFunc) {

	saveTicker := time.NewTicker(a.Config.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			a.FillMetric()

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func NewAgent(config *environment.AgentConfig) Agent {
	if config.StringTypeServer == constants.TypeSrvGRPC.String() {
		return newAgentGRPC(config)
	}

	return newAgentHTTP(config)
}

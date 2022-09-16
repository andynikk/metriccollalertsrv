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
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricsGauge map[string]repository.Gauge
type emtyArrMetrics []encoding.Metrics

type goRutine struct {
	ctx context.Context
	cnf context.CancelFunc
	wg  sync.WaitGroup
}

type data struct {
	mx           sync.RWMutex
	pollCount    int64
	metricsGauge MetricsGauge
}

type agent struct {
	cfg      environment.AgentConfig
	goRutine goRutine
	data     data
}

func (eam *emtyArrMetrics) prepareMetrics() ([]byte, error) {
	arrMetrics, err := json.MarshalIndent(eam, "", " ")
	if err != nil {
		return nil, err
	}

	gziparrMetrics, err := compression.Compress(arrMetrics)
	if err != nil {
		return nil, err
	}

	return gziparrMetrics, nil
}

func (a *agent) fillMetric(mems *runtime.MemStats) {

	a.data.mx.Lock()

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

	a.data.mx.Unlock()

}

func (a *agent) metrixOtherScan() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	saveTicker := time.NewTicker(a.cfg.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			fmt.Print("tack...\n")

			cpuUtilization, _ := cpu.Percent(2*time.Second, false)
			swapMemory, err := mem.SwapMemory()
			if err != nil {
				constants.Logger.ErrorLog(err)
			}
			CPUutilization1 := repository.Gauge(0)
			for _, val := range cpuUtilization {
				CPUutilization1 = repository.Gauge(val)
				break
			}
			a.data.mx.Lock()

			a.data.metricsGauge["TotalMemory"] = repository.Gauge(swapMemory.Total)
			a.data.metricsGauge["FreeMemory"] = repository.Gauge(swapMemory.Free)
			a.data.metricsGauge["CPUutilization1"] = CPUutilization1

			a.data.mx.Unlock()

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (a *agent) metrixScan() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	saveTicker := time.NewTicker(a.cfg.PollInterval)
	for {
		select {
		case <-saveTicker.C:

			fmt.Print("tick...\n")

			var mems runtime.MemStats
			runtime.ReadMemStats(&mems)
			a.fillMetric(&mems)

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (a *agent) old_metrixScan() {

	saveTicker := time.NewTicker(a.cfg.PollInterval)
	stop := make(chan bool, 1)

	for {
		select {
		case <-saveTicker.C:

			var mems runtime.MemStats
			runtime.ReadMemStats(&mems)
			a.fillMetric(&mems)

		case <-stop:
			a.goRutine.wg.Done()
			return
		}
	}
}

func (a *agent) Post2Server(allMterics []byte) error {

	addressPost := fmt.Sprintf("http://%s/updates", a.cfg.Address)
	req, err := http.NewRequest("POST", addressPost, bytes.NewReader(allMterics))
	if err != nil {

		constants.Logger.ErrorLog(err)

		return errors.New("-- ошибка отправки данных на сервер (1)")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
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

func (a *agent) goPost2Server(allMetrics emtyArrMetrics) {
	gziparrMetrics, err := allMetrics.prepareMetrics()
	if err != nil {
		constants.Logger.ErrorLog(err)
	}
	if err := a.Post2Server(gziparrMetrics); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func prepareMetrics(allMetrics *emtyArrMetrics) ([]byte, error) {

	arrMetrics, err := json.MarshalIndent(&allMetrics, "", " ")
	if err != nil {
		return nil, err
	}

	gziparrMetrics, err := compression.Compress(arrMetrics)
	if err != nil {
		return nil, err
	}
	return gziparrMetrics, nil
}

func (a *agent) MakeRequest() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	reportTicker := time.NewTicker(a.cfg.ReportInterval)

	for {
		select {
		case <-reportTicker.C:
			fmt.Print("toe...\n")
			allMetrics := make(emtyArrMetrics, 0)

			chanMatrics := make(chan encoding.Metrics, constants.ButchSize)

			//allMetrics := make(emtyArrMetrics, constants.ButchSize)
			//i := 0
			tempMetricsGauge := &a.data.metricsGauge
			for key, val := range *tempMetricsGauge {
				valFloat64 := float64(val)

				msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
				heshVal := cryptohash.HeshSHA256(msg, a.cfg.Key)

				metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
				chanMatrics <- metrica

				if cap(chanMatrics) != 0 && len(chanMatrics) == cap(chanMatrics) {
					for mt := range chanMatrics {
						allMetrics = append(allMetrics, mt)
						if len(chanMatrics) == 0 {
							break
						}
					}
					gziParrMterics, err := prepareMetrics(&allMetrics)
					if err != nil {
						constants.Logger.ErrorLog(err)
					}
					if err := a.Post2Server(gziParrMterics); err != nil {
						constants.Logger.ErrorLog(err)
					}
					allMetrics = make(emtyArrMetrics, 0)
				}

				cPollCount := repository.Counter(a.data.pollCount)
				msg = fmt.Sprintf("%s:counter:%d", "PollCount", a.data.pollCount)
				heshVal = cryptohash.HeshSHA256(msg, a.cfg.Key)
				chanMatrics <- encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.data.pollCount, Hash: heshVal}
				for mt := range chanMatrics {
					allMetrics = append(allMetrics, mt)
					if len(chanMatrics) == 0 {
						break
					}
				}
				gziparrMetrics, err := allMetrics.prepareMetrics()
				if err != nil {
					constants.Logger.ErrorLog(err)
				}
				if err := a.Post2Server(gziparrMetrics); err != nil {
					constants.Logger.ErrorLog(err)
				}
				//
				//	msg := fmt.Sprintf("%s:gauge:%f", key, valFloat64)
				//	heshVal := cryptohash.HeshSHA256(msg, a.cfg.Key)
				//
				//	metrica := encoding.Metrics{ID: key, MType: val.Type(), Value: &valFloat64, Hash: heshVal}
				//	allMetrics[i] = metrica
				//
				//	i++
				//	if i == constants.ButchSize {
				//		go a.goPost2Server(allMetrics)
				//		allMetrics = make(emtyArrMetrics, constants.ButchSize)
				//		i = 0
				//	}
				//}

				//cPollCount := repository.Counter(a.data.pollCount)
				//msg := fmt.Sprintf("%s:counter:%d", "PollCount", a.data.pollCount)
				//heshVal := cryptohash.HeshSHA256(msg, a.cfg.Key)
				//
				//metrica := encoding.Metrics{ID: "PollCount", MType: cPollCount.Type(), Delta: &a.data.pollCount, Hash: heshVal}
				//allMetrics[i+1] = metrica
				//go a.goPost2Server(allMetrics)
			}
		case <-ctx.Done():
			cancelFunc()
			//a.goRutine.wg.Done() //
			return
		}
	}
}

func main() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	agent := agent{
		cfg: environment.SetConfigAgent(),
		data: data{
			pollCount:    0,
			metricsGauge: make(MetricsGauge),
		},
		goRutine: goRutine{
			ctx: ctx,
			cnf: cancelFunc,
		},
	}

	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go agent.metrixScan()
	//agent.goRutine.wg.Add(1)

	go agent.metrixOtherScan()
	//agent.goRutine.wg.Add(1)
	//
	go agent.MakeRequest()
	//agent.goRutine.wg.Add(1)

	//agent.goRutine.wg.Wait()

	//stop <- true

	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	//<-stop

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func loadStoreMetrics(rs *handlers.RepStore, patch string) {

	res, err := ioutil.ReadFile(patch)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var arrMatric []encoding.Metrics
	if err := json.Unmarshal(res, &arrMatric); err != nil {
		fmt.Println(err.Error())
		return
	}

	rs.MX.Lock()
	defer rs.MX.Unlock()

	for _, val := range arrMatric {
		rs.AddNilMetric(val.MType, val.ID)
		rs.MutexRepo[val.ID].Set(val)
	}
	fmt.Println(rs.MutexRepo)

}

func SaveMetric2File(rs *handlers.RepStore, patch string) {

	rs.SaveMetric2File(patch)
}

func main() {

	//ctx, cancel := context.WithCancel(context.Background())
	//go handleSignals(cancel)
	//
	rs := handlers.NewRepStore()

	cfg := &handlers.Config{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	if cfg.RESTORE {
		loadStoreMetrics(rs, cfg.STORE_FILE)
	}

	s := &http.Server{
		Addr:    cfg.ADDRESS,
		Handler: rs.Router,
	}
	if s.ListenAndServe(); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	saveTicker := time.NewTicker(time.Duration(cfg.STORE_INTERVAL) * time.Second)

	for {
		select {
		case <-saveTicker.C:
			SaveMetric2File(rs, cfg.STORE_FILE)
			//case <-reportTicker.C:
			//	MakeRequest(metric)
		}
	}

	//if cfg.Restore {
	//	go loadStoreMetrics(rs, wg)
	//}
	//go SaveMetric2File(rs, cfg, wg)
	//
	//wg.Wait()
	//
	//for {
	//	select {
	//	case <-ctx.Done():
	//
	//		rs.SaveMetric2File(patch)
	//		log.Panicln("server stopped")
	//
	//	default:
	//
	//		timer := time.NewTimer(2 * time.Second)
	//		<-timer.C
	//	}
	//}

}

func handleSignals(cancel context.CancelFunc) {
	//sigCh := make(chan os.Signal)
	//signal.Notify(sigCh, os.Interrupt, os.Kill)
	//<-sigCh

	//for {
	//	sig := <-sigCh
	//	switch sig {
	//	case os.Interrupt:
	//		fmt.Println("canceled")
	//		cancel()
	//		return
	//	}
	//}
}

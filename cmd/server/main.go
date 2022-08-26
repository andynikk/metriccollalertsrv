package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
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
		rs.SetValueInMapJSON(val)
	}
	fmt.Println(rs.MutexRepo)

}

func SaveMetric2File(rs *handlers.RepStore, patch string, interval int64) {

	saveTicker := time.NewTicker(time.Duration(interval) * time.Second)

	for key := range saveTicker.C {
		fmt.Println(key)
		rs.SaveMetric2File(patch)
	}

	//for {
	//	select {
	//	case <-saveTicker.C:
	//		rs.SaveMetric2File(patch)
	//	default:
	//		fmt.Println("--")
	//	}
	//}
}

func main() {
	//fmt.Println("Запуск сервера")

	rs := handlers.NewRepStore()

	cfg := &handlers.Config{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	log.Println("Адрес сервера:", cfg.Address)

	if cfg.Restore {
		loadStoreMetrics(rs, cfg.StoreFile)
	}

	//handlers.AddrServ = os.Getenv("ADDRESS")
	//fmt.Println("AddrServ:", handlers.AddrServ)
	//if handlers.AddrServ == "" {
	//	handlers.AddrServ = "localhost:8080"
	//}

	go SaveMetric2File(rs, cfg.StoreFile, cfg.StoreInterval)

	go func() {
		s := &http.Server{
			Addr: cfg.Address,
			//Addr:    handlers.AddrServ,
			Handler: rs.Router}

		if err := s.ListenAndServe(); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
	}()
	//
	stop := make(chan os.Signal, 1024)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	rs.SaveMetric2File(cfg.StoreFile)
	log.Panicln("server stopped")

}

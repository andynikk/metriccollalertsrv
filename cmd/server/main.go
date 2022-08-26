package main

import (
	"encoding/json"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/caarlos0/env/v6"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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
	for {
		select {
		case <-saveTicker.C:
			rs.SaveMetric2File(patch)
		}
	}
}

func main() {
	fmt.Println("Запуск сервера")

	rs := handlers.NewRepStore()

	cfg := &handlers.Config{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	//if cfg.RESTORE {
	//	loadStoreMetrics(rs, cfg.STORE_FILE)
	//}

	handlers.AddrServ = os.Getenv("ADDRESS")
	fmt.Println("AddrServ:", handlers.AddrServ)
	if handlers.AddrServ == "" {
		handlers.AddrServ = "localhost:8080"
	}

	fmt.Println("Адрес сервера:", handlers.AddrServ)

	//go SaveMetric2File(rs, cfg.STORE_FILE, cfg.STORE_INTERVAL)

	go func() {
		//s := &http.Server{
		//	//Addr:    cfg.ADDRESS,
		//	Addr:    handlers.AddrServ,
		//	Handler: rs.Router}
		//
		//if err := s.ListenAndServe(); err != nil {
		//	fmt.Printf("%+v\n", err)
		//	return
		//}
		log.Fatal(http.ListenAndServe(":8080", rs.Router))
	}()
	//
	stop := make(chan os.Signal)
	//signal.Notify(stop, os.Interrupt, os.Kill)
	signal.Stop(stop)
	//<-stop
	//rs.SaveMetric2File(cfg.STORE_FILE)
	log.Panicln("server stopped")

}

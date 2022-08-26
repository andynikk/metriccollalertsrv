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

	//args := os.Args
	//fmt.Printf("All arguments: %v\n", args)
	//command := os.Args[0]
	//fmt.Printf("Command name: %v\n", command)
	//parameters := os.Args[1:]
	//fmt.Printf("Parameters: %v\n", parameters)

	fmt.Println("Run server")
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

	go SaveMetric2File(rs, cfg.STORE_FILE, cfg.STORE_INTERVAL)

	go func() {
		addServ := os.Getenv("ADDRESS")

		s := &http.Server{
			Addr: addServ,
			//Addr:    ":8080",
			Handler: rs.Router}

		if err := s.ListenAndServe(); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	rs.SaveMetric2File(cfg.STORE_FILE)
	log.Panicln("-server stopped")

}

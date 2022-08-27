package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func SaveMetric2File(rs *handlers.RepStore) {

	saveTicker := time.NewTicker(rs.Config.StoreInterval) // * time.Second)

	for key := range saveTicker.C {
		fmt.Println(key)
		rs.SaveMetric2File()
	}

	for {
		select {
		case <-saveTicker.C:
			rs.SaveMetric2File()
		default:
			fmt.Println("--")
		}
	}
}

func main() {

	rs := handlers.NewRepStore()

	if rs.Config.Restore {
		rs.LoadStoreMetrics()
	}

	go SaveMetric2File(rs)

	go func() {
		s := &http.Server{
			Addr:    rs.Config.Address,
			Handler: rs.Router}

		if err := s.ListenAndServe(); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
	}()

	stop := make(chan os.Signal, 1024)
	signal.Notify(stop, os.Interrupt) //, os.Kill)
	<-stop
	rs.SaveMetric2File()
	log.Panicln("server stopped")

}

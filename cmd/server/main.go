package main

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {

	rs := handlers.NewRepStore()

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
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Panicln("server stopped")

}

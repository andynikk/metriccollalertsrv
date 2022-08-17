package main

import (
	"log"
	"net/http"

	"github.com/andynikk/metriccollalertsrv/internal/consts"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func main() {

	log.Fatal(http.ListenAndServe(consts.PortServer, handlers.NewRepStore()))

}

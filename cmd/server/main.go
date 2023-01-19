package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func main() {

	config := environment.InitConfigServer()
	srv := handlers.NewServer(config)
	srv.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	srv.Shutdown()
}

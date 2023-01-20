package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func main() {
	fmt.Println("+++++++++++++++ start server", 1)
	config := environment.InitConfigServer()
	fmt.Println("+++++++++++++++ start server", 2)
	srv := handlers.NewServer(config)
	fmt.Println("+++++++++++++++ start server", 3)
	srv.Run()
	fmt.Println("+++++++++++++++ start server", 4)

	stop := make(chan os.Signal, 1)
	fmt.Println("+++++++++++++++ start server", 5)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	fmt.Println("+++++++++++++++ start server", 6)
	<-stop
	fmt.Println("+++++++++++++++ start server", 7)
	srv.Shutdown()
}

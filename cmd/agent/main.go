package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/metriccollalertsrv/internal/agent"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	config := environment.InitConfigAgent()

	a := agent.NewAgent(config)
	a.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop

	a.Stop()
}

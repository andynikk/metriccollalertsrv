package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

type server struct {
	storege handlers.RepStore
}

func Shutdown(rs *handlers.RepStore) {
	rs.Lock()
	defer rs.Unlock()

	for _, val := range rs.Config.TypeMetricsStorage {
		val.WriteMetric(rs.PrepareDataBU())
	}
	constants.Logger.InfoLog("server stopped")
}

func main() {

	server := new(server)
	handlers.NewRepStore(&server.storege)
	fmt.Println(server.storege.Config.Address)
	if server.storege.Config.Restore {
		go server.storege.RestoreData()
	}

	go server.storege.BackupData()

	go func() {
		s := &http.Server{
			Addr:    server.storege.Config.Address,
			Handler: server.storege.Router}

		if err := s.ListenAndServe(); err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	Shutdown(&server.storege)

}

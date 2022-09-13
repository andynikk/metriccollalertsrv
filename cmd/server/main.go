package main

import (
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	storege repository.StoreMetrics
}

func Shutdown(sm *repository.StoreMetrics) {
	for _, val := range sm.MapTypeStore {
		val.WriteMetric(sm.PrepareDataBU())
	}
	constants.Logger.InfoLog("server stopped")
}

func main() {

	rs := new(handlers.RepStore)
	server := new(server)

	handlers.NewRepStore(rs, &server.storege)

	if rs.Config.Restore {
		server.storege.RestoreData()
	}

	go server.storege.BackupData()

	go func() {
		s := &http.Server{
			Addr:    rs.Config.Address,
			Handler: rs.Router}

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

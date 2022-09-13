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
	storege handlers.RepStore
}

func Shutdown(sm *repository.StoreMetrics) {
	for _, val := range sm.MapTypeStore {
		val.WriteMetric(sm.PrepareDataBU())
	}
	constants.Logger.InfoLog("server stopped")
}

func main() {

	//rs := new(handlers.RepStore)
	server := new(server)

	handlers.NewRepStore(&server.storege)

	if server.storege.Config.Restore {
		server.storege.MutexRepo.RestoreData()
	}

	//go server.storege.BackupData()
	go server.storege.MutexRepo.RestoreData()

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
	//Shutdown(&server.storege)
	Shutdown(server.storege.MutexRepo)

}

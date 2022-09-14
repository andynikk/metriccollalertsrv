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
	Storege repository.StoreMetrics `json:"storege"`
}

func Shutdown(sm *repository.StoreMetrics) {
	for _, val := range sm.MapTypeStore {
		val.WriteMetric(sm.PrepareDataBU())
	}
	constants.Logger.InfoLog("server stopped")
}

func main() {

	rs := handlers.NewRepStore()
	server := server{
		Storege: repository.StoreMetrics{
			MapTypeStore:  rs.MutexRepo.MapTypeStore,
			HashKey:       rs.MutexRepo.HashKey,
			StoreInterval: rs.MutexRepo.StoreInterval,
			MX:            rs.MutexRepo.MX,
			Repo:          rs.MutexRepo.Repo,
		},
	}

	if rs.Config.Restore {
		server.Storege.RestoreData()
	}

	//ctx, cancel := context.WithCancel(context.Background())
	//go server.Storege.BackupData(ctx, cancel)
	go server.Storege.BackupData()

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
	Shutdown(&server.Storege)

}

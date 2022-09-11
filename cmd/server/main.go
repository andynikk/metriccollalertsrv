package main

import (
	"context"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func BackupData(rs *handlers.RepStore, ctx context.Context, cancel context.CancelFunc) {

	saveTicker := time.NewTicker(rs.Config.StoreInterval)

	for {
		select {
		case <-saveTicker.C:
			rs.PrepareDataBU()
			rs.SaveMetric()
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func Shutdown(rs *handlers.RepStore) {
	rs.PrepareDataBU()
	rs.SaveMetric()
	rs.Logger.InfoLog("server stopped")
}

func main() {

	rs := handlers.NewRepStore()
	if rs.Config.Restore {
		switch rs.Config.TypeMetricsStorage {
		case constants.MetricsStorageDB:
			rs.LoadStoreMetricsFromDB()
		case constants.MetricsStorageFile:
			rs.LoadStoreMetricsFromFile()
		}
	}
	ctx, cancel := context.WithCancel(rs.Ctx)
	go BackupData(rs, ctx, cancel)

	go func() {
		s := &http.Server{
			Addr:    rs.Config.Address,
			Handler: rs.Router}

		if err := s.ListenAndServe(); err != nil {
			constants.Logger.Error().Err(err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	Shutdown(rs)

}

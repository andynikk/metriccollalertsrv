package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func BackupData(rs *handlers.RepStore, ctx context.Context, cancel context.CancelFunc) {

	saveTicker := time.NewTicker(rs.Config.StoreInterval)

	for {
		select {
		case <-saveTicker.C:
			rs.SaveMetric()
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func Shutdown(rs *handlers.RepStore) {
	rs.SaveMetric()
	rs.Logger.InfoLog("server stopped")
}

func main() {

	rs := handlers.NewRepStore()

	if rs.Config.Restore {
		switch rs.Config.TypeMetricsStorage {
		case constants.MetricsStorageDb:
			rs.LoadStoreMetricsFromDB()
		case constants.MetricsStorageFile:
			rs.LoadStoreMetricsFromFile()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	go BackupData(rs, ctx, cancel)

	fmt.Println("******", rs.Config.Address)
	go func() {
		s := &http.Server{
			Addr:    rs.Config.Address,
			Handler: rs.Router}
		fmt.Println("******", s.Addr)

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

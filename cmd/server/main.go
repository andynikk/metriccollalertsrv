package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func LoadData(rs *handlers.RepStore) {
	for _, val := range rs.Config.TypeMetricsStorage {
		arrMatric, err := val.GetMetric()
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		rs.MX.Lock()
		defer rs.MX.Unlock()

		for _, val := range arrMatric {
			rs.SetValueInMapJSON(val)
		}
	}
}

func BackupData(rs *handlers.RepStore, ctx context.Context, cancel context.CancelFunc) {

	saveTicker := time.NewTicker(rs.Config.StoreInterval)
	for {
		select {
		case <-saveTicker.C:
			for _, val := range rs.Config.TypeMetricsStorage {
				val.WriteMetric(rs.PrepareDataBU())
			}
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func Shutdown(rs *handlers.RepStore) {
	for _, val := range rs.Config.TypeMetricsStorage {
		val.WriteMetric(rs.PrepareDataBU())
	}
	constants.Logger.InfoLog("server stopped")
}

func main() {

	rs := handlers.NewRepStore()
	if rs.Config.Restore {
		LoadData(rs)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go BackupData(rs, ctx, cancel)

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
	Shutdown(rs)

}

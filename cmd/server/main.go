package main

import (
	"context"
	"fmt"
	"log"
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
			rs.SaveMetric2File()
		case <-ctx.Done():
			cancel()
			return
		default:

		}
	}
}

func main() {

	rs := handlers.NewRepStore()

	if rs.Config.Restore {
		rs.LoadStoreMetrics()
	}

	ctx, cancel := context.WithCancel(context.Background())
	go BackupData(rs, ctx, cancel)

	go func() {
		s := &http.Server{
			Addr:    rs.Config.Address,
			Handler: rs.Router}

		if err := s.ListenAndServe(); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	rs.SaveMetric2File()
	log.Panicln("server stopped")

}

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
)

func main() {

	config := environment.InitConfigServer()
	srv := handlers.NewServer(config)
	go srv.RestoreData()
	go srv.BackupData()

	go func() {
		err := srv.Start()
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-stop
	srv.Shutdown()

	//server := new(server)
	//handlers.NewRepStore(&server.storege)
	//fmt.Println(server.storege.Config.Address)
	//if server.storege.Config.Restore {
	//	go server.storege.RestoreData()
	//}
	//
	//go server.storege.BackupData()
	//
	//go func() {
	//	s := &http.Server{
	//		Addr:    server.storege.Config.Address,
	//		Handler: server.storege.Router}
	//
	//	if err := s.ListenAndServe(); err != nil {
	//		constants.Logger.ErrorLog(err)
	//		return
	//	}
	//}()
	//
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	//<-stop
	//Shutdown(&server.storege)

}

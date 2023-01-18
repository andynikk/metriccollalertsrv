package handlers

import (
	"net"
	"net/http"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/middlware"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type KeyContext string

type serverHTTP struct {
	*RepStore
	chi.Router
}

type serverGRPS struct {
	*RepStore
}

type HServer interface {
}

type Server interface {
	Start() error
	RestoreData()
	BackupData()
	Shutdown()
	GetRouter() chi.Router
	GetRepStore() *RepStore
}

func (s *serverGRPS) GetRepStore() *RepStore {
	return s.RepStore
}

func (s *serverHTTP) GetRepStore() *RepStore {
	return s.RepStore
}

func (s *serverGRPS) GetRouter() chi.Router {
	return nil
}

func (s *serverHTTP) GetRouter() chi.Router {
	return s.Router
}

func (s *serverHTTP) Start() error {
	HTTPServer := &http.Server{
		Addr:    s.Config.Address,
		Handler: s.GetRouter(),
	}

	if err := HTTPServer.ListenAndServe(); err != nil {
		constants.Logger.ErrorLog(err)
		return err
	}
	return nil
}

func (s *serverGRPS) Start() error {

	server := grpc.NewServer(middlware.WithServerUnaryInterceptor())
	srv := &serverGRPS{
		RepStore: s.RepStore,
	}

	RegisterMetricCollectorServer(server, srv)
	l, err := net.Listen("tcp", constants.AddressServer)
	if err != nil {
		return err
	}

	if err = server.Serve(l); err != nil {
		return err
	}

	return nil
}

func (s *serverGRPS) RestoreData() {
	if s.Config.Restore {
		s.RepStore.RestoreData()
	}
}

func (s *serverGRPS) BackupData() {
	s.RepStore.BackupData()
}

func (s *serverHTTP) Shutdown() {
	s.RepStore.Shutdown()
}

func (s *serverGRPS) Shutdown() {
	s.RepStore.Shutdown()
}

func newHTTPServer(configServer *environment.ServerConfig) *serverHTTP {

	server := new(serverHTTP)

	server.Config = *configServer
	server.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	server.MutexRepo = make(repository.MutexRepo)
	NewRepStore(server)

	return server
}

func newGRPCServer(configServer *environment.ServerConfig) *serverGRPS {
	server := new(serverGRPS)
	server.RepStore = &RepStore{}
	server.RepStore.Config = *configServer
	server.RepStore.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	server.RepStore.MutexRepo = make(repository.MutexRepo)
	return server
}

// NewServer реализует фабричный метод.
func NewServer(configServer *environment.ServerConfig) Server {
	if configServer.TypeServer == constants.TypeSrvGRPC.String() {
		return newGRPCServer(configServer)
	}

	return newHTTPServer(configServer)
}

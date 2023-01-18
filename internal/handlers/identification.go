package handlers

import (
	"net/http"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
)

type serverHTTP struct {
	RepStore
	chi.Router
}

type serverGRPS struct {
	RepStore
}

type HServer interface {
}

type Server interface {
	Start() error
	RestoreData()
	BackupData()
	Shutdown()
	GetRouter() chi.Router
	GetRepStore() RepStore
}

func (s *serverGRPS) GetRepStore() RepStore {
	return s.RepStore
}

func (s *serverHTTP) GetRepStore() RepStore {
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
		Handler: s.Router,
	}

	if err := HTTPServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *serverGRPS) Start() error {

	//server := grpc.NewServer(middlware.WithServerUnaryInterceptor())
	//RegisterMetricCollectorServer(server, s)
	//l, err := net.Listen("tcp", constants.AddressServer)
	//if err != nil {
	//	return err
	//}
	//
	//if err = server.Serve(l); err != nil {
	//	return err
	//}
	//
	return nil
}

func (s *serverGRPS) RestoreData() {
	if s.Config.Restore {
		s.RestoreData()
	}
}

func (s *serverGRPS) BackupData() {
	s.BackupData()
}

func (s *serverHTTP) Shutdown() {
	s.RepStore.Shutdown()
}

func (s *serverGRPS) Shutdown() {
	s.RepStore.Shutdown()
}

func (rs *RepStore) Shutdown() {
	rs.Lock()
	defer rs.Unlock()

	for _, val := range rs.Config.StorageType {
		val.WriteMetric(rs.PrepareDataBuckUp())
	}
	constants.Logger.InfoLog("server stopped")
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
	//
	//server.storage.Config = configServer
	//server.storage.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	//
	//grpchandlers.NewRepStore(&server.storage)
	//fmt.Println(server.storage.Config.Address)
	//
	//gRepStore := general.New[grpchandlers.RepStore]()
	//gRepStore.Set(constants.TypeSrvGRPC.String(), server.storage)
	//
	//srv := &GRPCServer{
	//	RepStore: gRepStore,
	//}
	//server.srv = *srv
	//
	return server
}

// NewServer реализует фабричный метод.
func NewServer(configServer *environment.ServerConfig) Server {
	if configServer.TypeServer == constants.TypeSrvGRPC.String() {
		return newGRPCServer(configServer)
	}

	return newHTTPServer(configServer)
}

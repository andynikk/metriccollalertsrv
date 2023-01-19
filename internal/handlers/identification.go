package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type KeyContext string

type ServerHTTP struct {
	*RepStore
	chi.Router
}

type serverGRPS struct {
	*RepStore
}

type Server interface {
	Run()
	RestoreData()
	BackupData()
	Shutdown()
	GetRepStore() *RepStore
}

func (s *serverGRPS) GetRepStore() *RepStore {
	return s.RepStore
}

func (s *ServerHTTP) GetRepStore() *RepStore {
	return s.RepStore
}

func (s *ServerHTTP) Run() {

	go s.RestoreData()
	go s.BackupData()

	go func() {
		HTTPServer := &http.Server{
			Addr:    s.Config.Address,
			Handler: s.Router,
		}

		if err := HTTPServer.ListenAndServe(); err != nil {
			constants.Logger.ErrorLog(err)
			return
		}
	}()
}

func (s *serverGRPS) Run() {

	go s.RestoreData()
	go s.BackupData()

	go func() {
		server := grpc.NewServer(s.WithServerUnaryInterceptor())
		srv := &serverGRPS{
			RepStore: s.RepStore,
		}

		RegisterMetricCollectorServer(server, srv)
		l, err := net.Listen("tcp", constants.AddressServer)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return
		}

		if err = server.Serve(l); err != nil {
			return
		}
	}()
}

func (s *serverGRPS) RestoreData() {
	if s.Config.Restore {
		s.RepStore.RestoreData()
	}
}

func (s *serverGRPS) BackupData() {
	s.RepStore.BackupData()
}

func (s *ServerHTTP) Shutdown() {
	s.RepStore.Shutdown()
}

func (s *serverGRPS) Shutdown() {
	s.RepStore.Shutdown()
}

func newHTTPServer(configServer *environment.ServerConfig) *ServerHTTP {

	server := new(ServerHTTP)
	server.RepStore = &RepStore{}
	server.Config = *configServer
	server.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	server.MutexRepo = make(repository.MutexRepo)
	NewRepStore(server)

	return server
}

func newGRPCServer(configServer *environment.ServerConfig) *serverGRPS {

	server := new(serverGRPS)
	server.RepStore = &RepStore{}
	server.Config = *configServer
	server.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	server.MutexRepo = make(repository.MutexRepo)

	return server
}

// NewServer реализует фабричный метод.
func NewServer(configServer *environment.ServerConfig) Server {
	if configServer.TypeServer == constants.TypeSrvGRPC.String() {
		return newGRPCServer(configServer)
	}

	return newHTTPServer(configServer)
}

func (s *ServerHTTP) ChiCheckIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		xRealIP := r.Header.Get("X-Real-IP")
		if xRealIP == "" {
			next.ServeHTTP(w, r)
			return
		}

		ok := networks.AddressAllowed(strings.Split(xRealIP, constants.SepIPAddress), s.Config.TrustedSubnet)
		if !ok {
			w.WriteHeader(errs.StatusHTTP(errs.ErrForbidden))
			return
		}

		next.ServeHTTP(w, r)
	})
}

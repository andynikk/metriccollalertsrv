package handlers

import (
	"net"
	"net/http"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	pb "github.com/andynikk/metriccollalertsrv/internal/pb"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type KeyContext string

type ServerHTTP struct {
	*RepStore
	chi.Router
}

type ServerGRPS struct {
	*RepStore
	pb.UnimplementedMetricCollectorServer
}

type Server interface {
	Run()
	RestoreData()
	BackupData()
	Shutdown()
	GetRepStore() *RepStore
}

func (s *ServerGRPS) GetRepStore() *RepStore {
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

func (s *ServerGRPS) Run() {

	go s.RestoreData()
	go s.BackupData()

	go func() {
		interceptors := []grpc.UnaryServerInterceptor{
			s.ServerInterceptor,
		}
		if s.Config.TrustedSubnet != nil {
			interceptors = append(interceptors, s.ServerInterceptor)
		}

		//server := grpc.NewServer(s.WithServerUnaryInterceptor())
		server := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))

		pb.RegisterMetricCollectorServer(server, s)
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

func (s *ServerGRPS) RestoreData() {
	if s.Config.Restore {
		s.RepStore.RestoreData()
	}
}

func (s *ServerGRPS) BackupData() {
	s.RepStore.BackupData()
}

func (s *ServerHTTP) Shutdown() {
	s.RepStore.Shutdown()
}

func (s *ServerGRPS) Shutdown() {
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

func newGRPCServer(configServer *environment.ServerConfig) *ServerGRPS {

	server := new(ServerGRPS)
	server.RepStore = &RepStore{}
	server.Config = *configServer
	server.PK, _ = encryption.InitPrivateKey(configServer.CryptoKey)
	server.MutexRepo = make(repository.MutexRepo)
	server.UnimplementedMetricCollectorServer = pb.UnimplementedMetricCollectorServer{}

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

		if s.Config.TrustedSubnet == nil {
			next.ServeHTTP(w, r)
			return
		}

		xRealIP := r.Header.Get("X-Real-IP")
		if xRealIP == "" {
			w.WriteHeader(errs.StatusHTTP(errs.ErrForbidden))
			return
		}

		allowed := networks.AddressAllowed(xRealIP, s.Config.TrustedSubnet)
		if !allowed {
			w.WriteHeader(errs.StatusHTTP(errs.ErrForbidden))
			return
		}

		next.ServeHTTP(w, r)
	})
}

package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	"github.com/andynikk/metriccollalertsrv/internal/pb"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerGRPS) MustEmbedUnimplementedMetricCollectorServer() {
	//TODO implement me
	panic("implement me")
}

func (s *ServerGRPS) UpdatesAllMetricsJSON(ctx context.Context, req *pb.RequestListMetrics) (*emptypb.Empty, error) {

	var storedData encoding.ArrMetrics
	s.Lock()
	defer s.Unlock()
	for _, val := range req.Metrics {
		metrics := encoding.Metrics{ID: val.Id, MType: val.Mtype, Value: val.Value, Delta: val.Delta, Hash: val.Hash}
		err := s.SetValueInMapJSON(metrics)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return &emptypb.Empty{}, err
		}
		s.MutexRepo[val.Id].GetMetrics(val.Mtype, val.Id, s.Config.Key)

		storedData = append(storedData, metrics)
	}
	s.Config.Storage.WriteMetric(storedData)
	return &emptypb.Empty{}, nil
}

func (s *ServerGRPS) UpdateOneMetricsJSON(ctx context.Context, req *pb.RequestMetrics) (*emptypb.Empty, error) {

	metrics := encoding.Metrics{ID: req.Metrics.Id, MType: req.Metrics.Mtype, Value: req.Metrics.Value,
		Delta: req.Metrics.Delta, Hash: req.Metrics.Hash}

	err := s.SetValueInMapJSON(metrics)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	mt := s.MutexRepo[metrics.ID].GetMetrics(metrics.MType, metrics.ID, s.Config.Key)

	var arrMetrics encoding.ArrMetrics
	arrMetrics = append(arrMetrics, mt)

	s.Config.Storage.WriteMetric(arrMetrics)
	return &emptypb.Empty{}, nil
}

func (s *ServerGRPS) UpdateOneMetrics(ctx context.Context, req *pb.RequestMetricsString) (*emptypb.Empty, error) {

	rp := s.GetRepStore()
	err := rp.setValueInMap(req.MetricsType, req.MetricsName, req.MetricsValue)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ServerGRPS) PingDataBase(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {

	if !repository.ConnDB(s.Config.Storage) {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return &emptypb.Empty{}, errs.ErrStatusInternalServer
	}

	return &emptypb.Empty{}, nil
}

func (s *ServerGRPS) GetValue(ctx context.Context, req *pb.RequestMetricsName) (*pb.ResponseString, error) {

	//strMetric, err := HandlerGetValue(req.MetName, s.RepStore)
	s.Lock()
	defer s.Unlock()

	metName := req.MetricsName
	if _, findKey := s.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d", 3))
		return nil, errs.ErrNotFound
	}

	strMetric := s.MutexRepo[metName].String()
	return &pb.ResponseString{Result: strMetric}, nil
}

func (s *ServerGRPS) GetValueJSON(ctx context.Context, req *pb.RequestGetMetrics) (*pb.ResponseMetrics, error) {

	v := encoding.Metrics{ID: req.Metrics.Id, MType: req.Metrics.Mtype.String()}
	metType := v.MType
	metName := v.ID

	s.Lock()
	defer s.Unlock()

	if _, findKey := s.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d %s %d %s", 1, metName, len(s.MutexRepo), s.Config.DatabaseDsn))
		return nil, errs.ErrNotFound
	}

	mt := s.MutexRepo[metName].GetMetrics(metType, metName, s.Config.Key)
	metricsJSON := &pb.Metrics{Id: mt.ID, Mtype: mt.MType, Value: mt.Value, Delta: mt.Delta, Hash: mt.Hash}
	return &pb.ResponseMetrics{Metrics: metricsJSON}, nil

}

func (s *ServerGRPS) GetListMetrics(ctx context.Context, req *emptypb.Empty) (*pb.ResponseListMetrics, error) {

	var arrM []*pb.Metrics
	for key, val := range s.MutexRepo {
		data := val.GetMetrics(val.Type(), key, s.Config.Key)

		arrM = append(arrM, &pb.Metrics{Id: data.ID,
			Mtype: data.MType,
			Delta: data.Delta,
			Value: data.Value,
			Hash:  data.Hash})
	}
	return &pb.ResponseListMetrics{Metrics: arrM}, nil
}

func (s *ServerGRPS) WithServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.ServerInterceptor)
}

func (s *ServerGRPS) ServerInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if s.Config.TrustedSubnet == nil {
		h, _ := handler(ctx, req)
		return h, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil
	}
	xRealIP := md.Get(strings.ToLower("X-Real-IP"))[0]

	allowed := networks.AddressAllowed(xRealIP, s.Config.TrustedSubnet)

	if !allowed {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("X-Real-Ip is not found: %s", xRealIP))
	}

	h, _ := handler(ctx, req)
	return h, nil
}

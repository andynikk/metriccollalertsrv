package handlers

import (
	"context"
	"errors"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"google.golang.org/grpc/metadata"
)

type Header map[string]string

func FillHeader(ctx context.Context) Header {
	mHeader := make(Header)

	mdI, _ := metadata.FromIncomingContext(ctx)
	for key, valMD := range mdI {
		for _, val := range valMD {
			mHeader[key] = val
		}
	}

	mdO, _ := metadata.FromOutgoingContext(ctx)
	for key, valMD := range mdO {
		for _, val := range valMD {
			mHeader[key] = val
		}
	}

	return mHeader
}

func (s *serverGRPS) mustEmbedUnimplementedMetricCollectorServer() {
	//TODO implement me
	panic("implement me")
}

func (s *serverGRPS) UpdatesAllMetricsJSON(ctx context.Context, req *UpdatesRequest) (*TextErrResponse, error) {

	rp := s.GetRepStore()
	err := HandlerUpdatesMetricJSON(req.Body, rp)

	return &TextErrResponse{Result: []byte(err.Error())}, err
}

func (s *serverGRPS) UpdateOneMetricsJSON(ctx context.Context, req *UpdateStrRequest) (*TextErrResponse, error) {
	_, _, err := HandlerUpdateMetricJSON(req.Body, s.GetRepStore())
	return &TextErrResponse{Result: []byte(err.Error())}, err
}

func (s *serverGRPS) UpdateOneMetrics(ctx context.Context, req *UpdateRequest) (*TextErrResponse, error) {

	rp := s.GetRepStore()
	err := rp.setValueInMap(string(req.MetType), string(req.MetName), string(req.MetType))
	return &TextErrResponse{Result: []byte(err.Error())}, err
}

func (s *serverGRPS) PingDataBases(ctx context.Context, req *EmptyRequest) (*TextErrResponse, error) {
	mapTypeStore := s.Config.StorageType

	if _, findKey := mapTypeStore[constants.MetricsStorageDB.String()]; !findKey {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	if mapTypeStore[constants.MetricsStorageDB.String()].ConnDB() == nil {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	return &TextErrResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) GetValue(ctx context.Context, req *UpdatesRequest) (*StatusResponse, error) {

	//val, err := s.RepStore.HandlerGetValue(req.Body)
	return &StatusResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) GetValueJSON(ctx context.Context, req *UpdatesRequest) (*FullResponse, error) {
	//header := FillHeader(ctx)
	//
	//h, body, err := s.RepStore.HandlerValueMetricaJSON(header, req.Body)
	//ok := true
	//if err != nil {
	//	ok = false
	//}
	//
	//var hdr string
	//for k, v := range h {
	//	hdr += fmt.Sprintf("%s:%s\n", k, v)
	//}

	return &FullResponse{Header: []byte(""), Body: []byte(""), Result: false}, nil
}

func (s *serverGRPS) GetListMetrics(ctx context.Context, req *EmptyRequest) (*StatusResponse, error) {
	//header := FillHeader(ctx)
	//_, val := s.RepStore.HandlerGetAllMetrics(header)
	return &StatusResponse{Result: []byte("")}, nil
}

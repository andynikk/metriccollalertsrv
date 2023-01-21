package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Header map[string]string

func FillMetadata(ctx context.Context) Header {
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

func (s *serverGRPS) UpdatesAllMetricsJSON(ctx context.Context, req *RequestUpdateByte) (*EmptyAnswer, error) {

	header := FillMetadata(ctx)
	contentEncoding := header["content-encoding"]
	contentEncryption := header["content-encryption"]

	bytBody := req.Body

	if contentEncryption != "" {
		bytBodyRsaDecrypt, err := s.PK.RsaDecrypt(bytBody)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2.1 %s", err.Error()))
			return nil, err
		}
		bytBody = bytBodyRsaDecrypt
	}

	if strings.Contains(contentEncoding, "gzip") {
		bytBodyDecompress, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			return nil, err
		}
		bytBody = bytBodyDecompress
	}

	rp := s.GetRepStore()
	err := HandlerUpdatesMetricJSON(bytBody, rp)
	if err != nil {
		return nil, err
	}

	return &EmptyAnswer{}, nil
}

func (s *serverGRPS) UpdateOneMetricsJSON(ctx context.Context, req *RequestUpdateByte) (*ResponseMetrics, error) {
	header := FillMetadata(ctx)

	contentEncoding := header["content-encoding"]
	bytBody := req.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBodyDecompress, err := compression.Decompress(req.Body)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			return nil, err
		}
		bytBody = bytBodyDecompress
	}
	arrMetrics, err := HandlerUpdateMetricJSON(bytBody, s.GetRepStore())
	if err != nil {
		return nil, err
	}
	mGRPC := &MetricsGRPC{
		ID:    arrMetrics[0].ID,
		MType: arrMetrics[0].MType,
		Delta: arrMetrics[0].Delta,
		Value: arrMetrics[0].Value,
		Hash:  arrMetrics[0].Hash,
	}
	arrM := []*MetricsGRPC{mGRPC}
	return &ResponseMetrics{Metrics: arrM}, nil
}

func (s *serverGRPS) UpdateOneMetrics(ctx context.Context, req *ResponseProperties) (*EmptyAnswer, error) {

	rp := s.GetRepStore()
	err := rp.setValueInMap(string(req.MetType), string(req.MetName), string(req.MetValue))
	if err != nil {
		return nil, err
	}

	return &EmptyAnswer{}, nil
}

func (s *serverGRPS) PingDataBase(ctx context.Context, req *EmptyRequest) (*EmptyAnswer, error) {

	if s.Config.Storage.ConnDB() == nil {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	return &EmptyAnswer{}, nil
}

func (s *serverGRPS) GetValue(ctx context.Context, req *ResponseProperties) (*ResponseString, error) {

	strMetric, err := HandlerGetValue(req.MetName, s.RepStore)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}
	return &ResponseString{Result: strMetric}, nil
}

func (s *serverGRPS) GetValueJSON(ctx context.Context, req *RequestByte) (*ResponseHeaderMetrics, error) {

	headerIn := FillMetadata(ctx)
	headerOut, bodyOut, err := HandlerValueMetricaJSON(headerIn, req.Body, s.RepStore)
	if err != nil {
		return nil, err
	}

	var hGRPC []*HeaderGRPC
	for k, v := range headerOut {
		hGRPC = append(hGRPC, &HeaderGRPC{Key: k, Value: v})
		if k == "gzip" {
			dataDecompress, err := compression.Decompress(bodyOut)
			if err != nil {
				constants.Logger.ErrorLog(err)
				continue
			}
			bodyOut = dataDecompress
		}
	}

	m := encoding.Metrics{}
	bodyJSON := bytes.NewReader(bodyOut)
	err = json.NewDecoder(bodyJSON).Decode(&m)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}

	var mGRPC []*MetricsGRPC
	mGRPC = append(mGRPC, &MetricsGRPC{ID: m.ID, MType: m.MType, Value: m.Value, Delta: m.Delta, Hash: m.Hash})
	return &ResponseHeaderMetrics{Header: hGRPC, Metrics: mGRPC}, nil
}

func (s *serverGRPS) GetListMetrics(ctx context.Context, req *EmptyRequest) (*ResponseMetrics, error) {

	var arrM []*MetricsGRPC
	for key, val := range s.MutexRepo {
		data := val.GetMetrics(val.Type(), key, s.Config.Key)

		arrM = append(arrM, &MetricsGRPC{ID: data.ID,
			MType: data.MType,
			Delta: data.Delta,
			Value: data.Value,
			Hash:  data.Hash})
	}
	return &ResponseMetrics{Metrics: arrM}, nil
}

func (s *serverGRPS) WithServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.ServerInterceptor)
}

func (s *serverGRPS) ServerInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	if s.Config.TrustedSubnet == "" {
		h, _ := handler(ctx, req)
		return h, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil
	}
	xRealIP := md.Get(strings.ToLower("X-Real-IP"))[0]

	addressRanges := strings.Split(xRealIP, constants.SepIPAddress)
	allowed := networks.AddressAllowed(addressRanges, s.Config.TrustedSubnet)

	if !allowed {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("X-Real-Ip is not found: %s", xRealIP))
	}

	h, _ := handler(ctx, req)
	return h, nil
}

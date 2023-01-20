package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
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

func (s *serverGRPS) UpdatesAllMetricsJSON(ctx context.Context, req *UpdatesRequest) (*EmptyRequest, error) {

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

	return &EmptyRequest{}, nil
}

func (s *serverGRPS) UpdateOneMetricsJSON(ctx context.Context, req *UpdateStrRequest) (*EmptyRequest, error) {
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
	_, _, err := HandlerUpdateMetricJSON(bytBody, s.GetRepStore())
	if err != nil {
		return nil, err
	}

	return &EmptyRequest{}, nil
}

func (s *serverGRPS) UpdateOneMetrics(ctx context.Context, req *UpdateRequest) (*EmptyRequest, error) {

	rp := s.GetRepStore()
	err := rp.setValueInMap(string(req.MetType), string(req.MetName), string(req.MetValue))
	if err != nil {
		return nil, err
	}

	return &EmptyRequest{}, nil
}

func (s *serverGRPS) PingDataBase(ctx context.Context, req *EmptyRequest) (*EmptyRequest, error) {

	if s.Config.Storage.ConnDB() == nil {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	return &EmptyRequest{}, nil
}

func (s *serverGRPS) GetValue(ctx context.Context, req *UpdatesRequest) (*StatusResponse, error) {

	strMetric, err := HandlerGetValue(req.Body, s.RepStore)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}
	return &StatusResponse{Result: []byte(strMetric)}, nil
}

func (s *serverGRPS) GetValueJSON(ctx context.Context, req *UpdatesRequest) (*FullResponse, error) {

	headerIn := FillMetadata(ctx)
	headerOut, bodyOut, err := HandlerValueMetricaJSON(headerIn, req.Body, s.RepStore)
	if err != nil {
		return nil, err
	}

	var hdr string
	for k, v := range headerOut {
		hdr += fmt.Sprintf("%s:%s\n", k, v)
	}

	return &FullResponse{Header: []byte(hdr), Body: bodyOut}, nil
}

func (s *serverGRPS) GetListMetrics(ctx context.Context, req *EmptyRequest) (*StatusResponse, error) {
	h := FillMetadata(ctx)

	arrMetricsAndValue := s.RepStore.TextMetricsAndValue()

	var strMetrics string
	content := `<!DOCTYPE html>
				<html>
				<head>
  					<meta charset="UTF-8">
  					<title>МЕТРИКИ</title>
				</head>
				<body>
				<h1>МЕТРИКИ</h1>
				<ul>
				`
	for _, val := range arrMetricsAndValue {
		content = content + `<li><b>` + val + `</b></li>` + "\n"
		if strMetrics != "" {
			strMetrics = strMetrics + ";"
		}
		strMetrics = strMetrics + val
	}
	content = content + `</ul>
						</body>
						</html>`

	acceptEncoding := h["Accept-Encoding"]

	metricsHTML := []byte(content)
	byteMeterics := bytes.NewBuffer(metricsHTML).Bytes()
	compData, err := compression.Compress(byteMeterics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	HeaderResponse := Header{}

	var bodyBate []byte
	if strings.Contains(acceptEncoding, "gzip") {
		HeaderResponse["content-encoding"] = "gzip"
		bodyBate = compData
	} else {
		bodyBate = metricsHTML
	}

	HeaderResponse["content-type"] = "text/html"
	HeaderResponse["metrics-val"] = strMetrics

	return &StatusResponse{Result: bodyBate}, nil
}

func (s *serverGRPS) WithServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.ServerInterceptor)
}

func (s *serverGRPS) ServerInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

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

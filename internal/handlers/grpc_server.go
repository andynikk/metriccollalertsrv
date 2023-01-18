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

	header := FillHeader(ctx)
	contentEncoding := header["content-encoding"]
	bytBody := req.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBodyDecompress, err := compression.Decompress(req.Body)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			return &TextErrResponse{Result: []byte(err.Error())}, err
		}
		bytBody = bytBodyDecompress
	}

	rp := s.GetRepStore()
	err := HandlerUpdatesMetricJSON(bytBody, rp)
	if err != nil {
		return &TextErrResponse{Result: []byte(err.Error())}, err
	}

	return &TextErrResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) UpdateOneMetricsJSON(ctx context.Context, req *UpdateStrRequest) (*TextErrResponse, error) {
	header := FillHeader(ctx)
	contentEncoding := header["content-encoding"]
	bytBody := req.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBodyDecompress, err := compression.Decompress(req.Body)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			return &TextErrResponse{Result: []byte(err.Error())}, err
		}
		bytBody = bytBodyDecompress
	}
	_, _, err := HandlerUpdateMetricJSON(bytBody, s.GetRepStore())
	if err != nil {
		return &TextErrResponse{Result: []byte(err.Error())}, err
	}

	return &TextErrResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) UpdateOneMetrics(ctx context.Context, req *UpdateRequest) (*TextErrResponse, error) {

	rp := s.GetRepStore()
	err := rp.setValueInMap(string(req.MetType), string(req.MetName), string(req.MetValue))
	if err != nil {
		return &TextErrResponse{Result: []byte(err.Error())}, err
	}

	return &TextErrResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) PingDataBases(ctx context.Context, req *EmptyRequest) (*TextErrResponse, error) {

	mapTypeStore := s.Config.StorageType

	fmt.Println("+++++++++++000", len(mapTypeStore))
	if _, findKey := mapTypeStore[constants.MetricsStorageDB.String()]; !findKey {
		fmt.Println("+++++++++++001")
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	if mapTypeStore[constants.MetricsStorageDB.String()].ConnDB() == nil {
		fmt.Println("+++++++++++002")
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		return nil, errs.ErrStatusInternalServer
	}

	return &TextErrResponse{Result: []byte("")}, nil
}

func (s *serverGRPS) GetValue(ctx context.Context, req *UpdatesRequest) (*StatusResponse, error) {

	strMetric, err := HandlerGetValue(req.Body, &s.RepStore)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, err
	}
	return &StatusResponse{Result: []byte(strMetric)}, nil
}

func (s *serverGRPS) GetValueJSON(ctx context.Context, req *UpdatesRequest) (*FullResponse, error) {

	headerIn := FillHeader(ctx)
	headerOut, bodyOut, err := HandlerValueMetricaJSON(headerIn, req.Body, &s.RepStore)
	if err != nil {
		return nil, err
	}

	var hdr string
	for k, v := range headerOut {
		hdr += fmt.Sprintf("%s:%s\n", k, v)
	}

	return &FullResponse{Header: []byte(hdr), Body: bodyOut, Result: true}, nil
}

func (s *serverGRPS) GetListMetrics(ctx context.Context, req *EmptyRequest) (*StatusResponse, error) {
	h := FillHeader(ctx)

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

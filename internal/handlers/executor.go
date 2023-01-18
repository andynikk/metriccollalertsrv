package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

func HandlerUpdatesMetricJSON(body []byte, rs *RepStore) error {

	var storedData encoding.ArrMetrics
	if err := json.Unmarshal(body, &storedData); err != nil {
		constants.Logger.ErrorLog(err)
		return errs.ErrStatusInternalServer
	}

	rs.Lock()
	defer rs.Unlock()

	for _, val := range storedData {
		err := rs.SetValueInMapJSON(val)
		if err != nil {
			return err
		}
		rs.MutexRepo[val.ID].GetMetrics(val.MType, val.ID, rs.Config.Key)
	}

	for _, val := range rs.Config.StorageType {
		val.WriteMetric(storedData)
	}

	return nil
}

func HandlerUpdateMetricJSON(body []byte, rs *RepStore) (Header, []byte, error) {

	bodyJSON := bytes.NewReader(body)

	v := encoding.Metrics{}
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		return nil, nil, errs.ErrStatusInternalServer
	}

	rs.Lock()
	defer rs.Unlock()

	headerRequest := Header{}
	headerRequest["Content-Type"] = "application/json"
	err = rs.SetValueInMapJSON(v)
	if err != nil {
		return nil, nil, err
	}

	mt := rs.MutexRepo[v.ID].GetMetrics(v.MType, v.ID, rs.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 4 %s", err.Error()))
		return nil, nil, errs.ErrStatusInternalServer
	}

	var arrMetrics encoding.ArrMetrics
	arrMetrics = append(arrMetrics, mt)

	for _, val := range rs.Config.StorageType {
		val.WriteMetric(arrMetrics)
	}

	return headerRequest, metricsJSON, nil
}

func HandlerGetValue(body []byte, rs *RepStore) (string, error) {

	metName := string(body)

	rs.Lock()
	defer rs.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d", 3))
		return "", errs.ErrNotFound
	}

	strMetric := rs.MutexRepo[metName].String()
	return strMetric, nil

}

func HandlerValueMetricaJSON(header Header, body []byte, rs *RepStore) (Header, []byte, error) {

	var bodyJSON io.Reader
	bodyJSON = bytes.NewReader(body)

	acceptEncoding := header[strings.ToLower("Accept-Encoding")]
	contentEncoding := header[strings.ToLower("Content-Encoding")]

	if strings.Contains(contentEncoding, "gzip") {
		constants.Logger.InfoLog("-- метрика с агента gzip (value)")
		bytBody, err := io.ReadAll(bodyJSON)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return nil, nil, errs.ErrStatusInternalServer
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return nil, nil, errs.ErrStatusInternalServer
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	v := encoding.Metrics{}
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, nil, errs.ErrStatusInternalServer
	}
	metType := v.MType
	metName := v.ID

	rs.Lock()
	defer rs.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d %s %d %s", 1, metName, len(rs.MutexRepo), rs.Config.DatabaseDsn))
		return nil, nil, errs.ErrNotFound
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName, rs.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, nil, err
	}

	var byteMetrics []byte
	bt := bytes.NewBuffer(metricsJSON).Bytes()
	byteMetrics = append(byteMetrics, bt...)
	compData, err := compression.Compress(byteMetrics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	var bodyBate []byte

	headerOut := Header{}
	headerOut["Content-Type"] = "application/json"
	if strings.Contains(acceptEncoding, "gzip") {
		headerOut["Content-Encoding"] = "gzip"
		bodyBate = compData
	} else {
		bodyBate = metricsJSON
	}

	return headerOut, bodyBate, nil
}

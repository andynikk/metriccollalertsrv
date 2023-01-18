package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

func HandlerUpdatesMetricJSON(header Header, body []byte, rs *RepStore) error {

	var storedData encoding.ArrMetrics
	if err := json.Unmarshal(body, &storedData); err != nil {
		constants.Logger.ErrorLog(err)
		return errs.ErrStatusInternalServer
	}

	rs.Lock()
	defer rs.Unlock()

	for _, val := range storedData {
		rs.SetValueInMapJSON(val)
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
	res := rs.SetValueInMapJSON(v)

	mt := rs.MutexRepo[v.ID].GetMetrics(v.MType, v.ID, rs.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 4 %s", err.Error()))
		return nil, nil, errs.ErrStatusInternalServer
	}

	if res == http.StatusOK {
		var arrMetrics encoding.ArrMetrics
		arrMetrics = append(arrMetrics, mt)

		for _, val := range rs.Config.StorageType {
			val.WriteMetric(arrMetrics)
		}
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

package handlers

import (
	"encoding/json"

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

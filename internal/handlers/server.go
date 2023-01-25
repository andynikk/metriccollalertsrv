package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants/errs"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricType int
type MetricError int

const (
	GaugeMetric MetricType = iota
	CounterMetric
)

type MutexRepStore struct {
	sync.Mutex
	repository.MapMetrics
}

type RepStore struct {
	Config environment.ServerConfig
	PK     *encryption.KeyEncryption
	MutexRepStore
}

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

func NewRepStore(s *ServerHTTP) {

	s.Router = chi.NewRouter()
	rs := s.RepStore

	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.StripSlashes)

	s.Router.Group(func(r chi.Router) {
		if s.Config.TrustedSubnet != nil {
			r.Use(s.ChiCheckIP)
		}
		r.Post("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetricaPOST) //+
		r.Post("/update", rs.HandlerUpdateMetricJSON)                              //+
		r.Post("/updates", rs.HandlerUpdatesMetricJSON)                            //+
	})

	s.Router.NotFound(rs.HandlerNotFound)
	s.Router.HandleFunc("/", rs.HandleFunc)

	s.Router.Get("/", rs.HandlerGetAllMetrics)

	s.Router.Get("/ping", rs.HandlerPingDB)                        //+
	s.Router.Get("/value/{metType}/{metName}", rs.HandlerGetValue) //+
	s.Router.Post("/value", rs.HandlerValueMetricaJSON)            //+

	s.Router.HandleFunc("/debug/pprof", pprof.Index)
	s.Router.HandleFunc("/debug/pprof/", pprof.Index)
	s.Router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.Router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.Router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.Router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.Router.Handle("/debug/block", pprof.Handler("block"))
	s.Router.Handle("/debug/goroutine", pprof.Handler("goroutine"))
	s.Router.Handle("/debug/heap", pprof.Handler("heap"))
	s.Router.Handle("/debug/threadcreate", pprof.Handler("threadcreate"))
	s.Router.Handle("/debug/allocs", pprof.Handler("allocs"))
	s.Router.Handle("/debug/mutex", pprof.Handler("mutex"))
	s.Router.Handle("/debug/mutex", pprof.Handler("mutex"))

}

func (rs *RepStore) setValueInMap(metType string, metName string, metValue string) error {

	switch metType {
	case GaugeMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return errs.ErrBadRequest
			}
		} else {

			valG := repository.Gauge(0)
			if ok := valG.SetFromText(metValue); !ok {
				return errs.ErrBadRequest
			}
			rs.MutexRepo[metName] = &valG
		}

	case CounterMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return errs.ErrBadRequest
			}
		} else {

			valC := repository.Counter(0)
			if ok := valC.SetFromText(metValue); !ok {
				return errs.ErrBadRequest
			}
			rs.MutexRepo[metName] = &valC
		}
	default:
		return errs.ErrNotImplemented
	}

	return nil
}

func (rs *RepStore) SetValueInMapJSON(v encoding.Metrics) error {

	var heshVal string

	switch v.MType {
	case GaugeMetric.String():
		var valValue float64
		valValue = *v.Value

		msg := fmt.Sprintf("%s:gauge:%f", v.ID, valValue)
		heshVal = cryptohash.HashSHA256(msg, rs.Config.Key)
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valG := repository.Gauge(0)
			rs.MutexRepo[v.ID] = &valG
		}
	case CounterMetric.String():

		var valDelta int64
		valDelta = *v.Delta

		msg := fmt.Sprintf("%s:counter:%d", v.ID, valDelta)
		heshVal = cryptohash.HashSHA256(msg, rs.Config.Key)
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valC := repository.Counter(0)
			rs.MutexRepo[v.ID] = &valC
		}
	default:
		return errs.ErrNotImplemented
	}

	hashAgent := []byte(v.Hash)
	hashServer := []byte(heshVal)

	hmacEqual := hmac.Equal(hashServer, hashAgent)

	if v.Hash != "" && !hmacEqual {
		constants.Logger.InfoLog(fmt.Sprintf("++ %s, %s %s - %s (%s)", v.ID, v.MType, v.Hash, heshVal, rs.Config.Key))
		//return http.StatusBadRequest
	}

	rs.MutexRepo[v.ID].Set(v)
	return nil

}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metName := chi.URLParam(rq, "metName")

	rs.Lock()
	defer rs.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d", 3))
		rw.WriteHeader(errs.StatusHTTP(errs.ErrNotFound))
	}
	strMetric := rs.MutexRepo[metName].String()

	_, err := io.WriteString(rw, strMetric)
	if err != nil {
		constants.Logger.ErrorLog(err)
		rw.WriteHeader(errs.StatusHTTP(err))
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) Shutdown() {
	rs.Lock()
	defer rs.Unlock()

	rs.Config.Storage.WriteMetric(rs.PrepareDataBuckUp())
	constants.Logger.InfoLog("server stopped")
}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	rs.Lock()
	defer rs.Unlock()

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	err := rs.setValueInMap(metType, metName, metValue)
	rw.WriteHeader(errs.StatusHTTP(err))
}

func (rs *RepStore) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	bytBody, err := io.ReadAll(rq.Body)
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 1 %s", err.Error()))
		http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
		return
	}

	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}
	}

	bodyJSON := bytes.NewReader(bytBody)

	v := encoding.Metrics{}
	err = json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		http.Error(rw, "Ошибка получения JSON", errs.StatusHTTP(errs.ErrStatusInternalServer))
		return
	}

	rs.Lock()
	defer rs.Unlock()

	err = rs.SetValueInMapJSON(v)
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		http.Error(rw, "Ошибка получения JSON", errs.StatusHTTP(err))
		return
	}

	mt := rs.MutexRepo[v.ID].GetMetrics(v.MType, v.ID, rs.Config.Key)

	var arrMetrics encoding.ArrMetrics
	arrMetrics = append(arrMetrics, mt)

	rs.Config.Storage.WriteMetric(arrMetrics)

	metricsJSON, err := arrMetrics[0].MarshalMetrica()
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		http.Error(rw, "Ошибка получения JSON", errs.StatusHTTP(err))
		return
	}

	if _, err = rw.Write(metricsJSON); err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 5 %s", err.Error()))
		rw.WriteHeader(errs.StatusHTTP(errs.ErrStatusInternalServer))
		return
	}
	rw.WriteHeader(errs.StatusHTTP(err))
}

func (rs *RepStore) HandlerUpdatesMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	bodyJSON, err := io.ReadAll(rq.Body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "ошибка на сервере", http.StatusInternalServerError)

	}

	contentEncryption := rq.Header.Get("Content-Encryption")
	if contentEncryption != "" {
		bytBodyRsaDecrypt, err := rs.PK.RsaDecrypt(bodyJSON)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2.1 %s", err.Error()))
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
		}
		bodyJSON = bytBodyRsaDecrypt
	}

	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		decompressBody, err := compression.Decompress(bodyJSON)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}
		bodyJSON = decompressBody
	}

	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
	}

	var storedData encoding.ArrMetrics
	if err = json.Unmarshal(bodyJSON, &storedData); err != nil {
		constants.Logger.ErrorLog(err)
		rw.WriteHeader(errs.StatusHTTP(errs.ErrStatusInternalServer))
		return
	}

	rs.Lock()
	defer rs.Unlock()

	for _, val := range storedData {
		err := rs.SetValueInMapJSON(val)
		if err != nil {
			rw.WriteHeader(errs.StatusHTTP(err))
			return
		}
		rs.MutexRepo[val.ID].GetMetrics(val.MType, val.ID, rs.Config.Key)
	}

	rs.Config.Storage.WriteMetric(storedData)
	rw.WriteHeader(errs.StatusHTTP(err))
}

func (rs *RepStore) HandlerValueMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	bytBody, err := io.ReadAll(rq.Body)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
		return
	}

	var bodyJSON io.Reader
	bodyJSON = bytes.NewReader(bytBody)

	contentEncoding := rq.Header.Get(strings.ToLower("Content-Encoding"))
	if strings.Contains(contentEncoding, "gzip") {
		constants.Logger.InfoLog("-- метрика с агента gzip (value)")
		bytBody, err = io.ReadAll(bodyJSON)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	v := encoding.Metrics{}
	err = json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	rs.Lock()
	defer rs.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d %s %d %s", 1, metName, len(rs.MutexRepo), rs.Config.DatabaseDsn))
		http.Error(rw, "Ошибка получения Content-Encoding", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName, rs.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
		return
	}

	var byteMetrics []byte
	bt := bytes.NewBuffer(metricsJSON).Bytes()
	byteMetrics = append(byteMetrics, bt...)
	compData, err := compression.Compress(byteMetrics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	rw.Header().Add("Content-Type", "application/json")
	bodyBate := metricsJSON

	acceptEncoding := rq.Header.Get(strings.ToLower("Accept-Encoding"))
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Add("Content-Encoding", "gzip")
		bodyBate = compData
	}

	if _, err = rw.Write(bodyBate); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerPingDB(rw http.ResponseWriter, rq *http.Request) {
	defer rq.Body.Close()

	if rs.Config.Storage == nil {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := rs.Config.Storage.ConnDB()
	if err != nil {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	arrMetricsAndValue := rs.MapMetrics.TextMetricsAndValue()

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

	acceptEncoding := rq.Header.Get("Accept-Encoding")

	metricsHTML := []byte(content)
	byteMetrics := bytes.NewBuffer(metricsHTML).Bytes()
	compData, err := compression.Compress(byteMetrics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	bodyBate := metricsHTML
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Add("Content-Encoding", "gzip")
		bodyBate = compData
	}

	rw.Header().Add("Content-Type", "text/html")
	if _, err = rw.Write(bodyBate); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) PrepareDataBuckUp() encoding.ArrMetrics {

	var storedData encoding.ArrMetrics
	for key, val := range rs.MutexRepo {
		storedData = append(storedData, val.GetMetrics(val.Type(), key, rs.Config.Key))
	}
	return storedData
}

func (rs *RepStore) RestoreData() {
	rs.Lock()
	defer rs.Unlock()

	arrMetrics, err := rs.Config.Storage.GetMetric()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	for _, val := range arrMetrics {
		if err = rs.SetValueInMapJSON(val); err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
	}
}

func (rs *RepStore) BackupData() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	saveTicker := time.NewTicker(rs.Config.StoreInterval)
	for {
		select {
		case <-saveTicker.C:

			rs.Config.Storage.WriteMetric(rs.PrepareDataBuckUp())

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (rs *RepStore) HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Метрика "+r.URL.Path+" не найдена", http.StatusNotFound)

}

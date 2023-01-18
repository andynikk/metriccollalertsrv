package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/middlware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
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

type HTMLParam struct {
	Title       string
	TextMetrics []string
}

type RepStore struct {
	Config environment.ServerConfig
	PK     *encryption.KeyEncryption
	sync.Mutex
	repository.MapMetrics
}

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

func UserContextBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		//var user User
		//
		//err := json.NewDecoder(r.Body).Decode(&user)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//
		//ctx := context.WithValue(r.Context(), "user", user)
		//next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func NewRepStore(s *serverHTTP) {

	s.Router = chi.NewRouter()

	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.StripSlashes)

	s.Router.Use(middlware.ChiCheckIP)

	s.Router.NotFound(s.HandlerNotFound)
	s.Router.HandleFunc("/", s.HandleFunc)

	s.Router.Post("/update/{metType}/{metName}/{metValue}", s.HandlerSetMetricaPOST)
	s.Router.Post("/update", s.HandlerUpdateMetricJSON)
	s.Router.Post("/updates", s.HandlerUpdatesMetricJSON)

	s.Router.Get("/", s.HandlerGetAllMetrics)
	s.Router.Get("/value/{metType}/{metName}", s.HandlerGetValue)
	s.Router.Post("/value", s.HandlerValueMetricaJSON)
	s.Router.Get("/ping", s.HandlerPingDB)
}

func (s *serverHTTP) setValueInMap(metType string, metName string, metValue string) int {

	switch metType {
	case GaugeMetric.String():
		if val, findKey := s.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valG := repository.Gauge(0)
			if ok := valG.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			s.MutexRepo[metName] = &valG
		}

	case CounterMetric.String():
		if val, findKey := s.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valC := repository.Counter(0)
			if ok := valC.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			s.MutexRepo[metName] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	return http.StatusOK
}

func (s *serverHTTP) SetValueInMapJSON(v encoding.Metrics) int {

	var heshVal string

	switch v.MType {
	case GaugeMetric.String():
		var valValue float64
		valValue = *v.Value

		msg := fmt.Sprintf("%s:gauge:%f", v.ID, valValue)
		heshVal = cryptohash.HashSHA256(msg, s.Config.Key)
		if _, findKey := s.MutexRepo[v.ID]; !findKey {
			valG := repository.Gauge(0)
			s.MutexRepo[v.ID] = &valG
		}
	case CounterMetric.String():
		var valDelta int64
		valDelta = *v.Delta

		msg := fmt.Sprintf("%s:counter:%d", v.ID, valDelta)
		heshVal = cryptohash.HashSHA256(msg, s.Config.Key)
		if _, findKey := s.MutexRepo[v.ID]; !findKey {
			valC := repository.Counter(0)
			s.MutexRepo[v.ID] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	heshAgent := []byte(v.Hash)
	heshServer := []byte(heshVal)

	hmacEqual := hmac.Equal(heshServer, heshAgent)

	constants.Logger.InfoLog(fmt.Sprintf("-- %s - %s", v.Hash, heshVal))

	if v.Hash != "" && !hmacEqual {
		constants.Logger.InfoLog(fmt.Sprintf("++ %s - %s", v.Hash, heshVal))
		return http.StatusBadRequest
	}
	constants.Logger.InfoLog(fmt.Sprintf("** %s %s %v %d", v.ID, v.MType, v.Value, v.Delta))

	s.MutexRepo[v.ID].Set(v)
	return http.StatusOK

}

func (s *serverHTTP) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	fmt.Println("------- HandlerGetValue", metType, metName)

	s.Lock()
	defer s.Unlock()

	if _, findKey := s.MutexRepo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d", 3))
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	strMetric := s.MutexRepo[metName].String()
	_, err := io.WriteString(rw, strMetric)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (s *serverHTTP) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	IPAddressAllowed := rq.Context().Value(middlware.KeyValueContext("IP-Address-Allowed"))
	if IPAddressAllowed == "false" {
		return
	}

	s.Lock()
	defer s.Unlock()

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	rw.WriteHeader(s.setValueInMap(metType, metName, metValue))
}

func (s *serverHTTP) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	IPAddressAllowed := rq.Context().Value(middlware.KeyValueContext("IP-Address-Allowed"))
	if IPAddressAllowed == "false" {
		return
	}

	var bodyJSON io.Reader

	contentEncoding := rq.Header.Get("Content-Encoding")
	bodyJSON = rq.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(rq.Body)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 1 %s", err.Error()))
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.InfoLog(fmt.Sprintf("$$ 2 %s", err.Error()))
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	v := encoding.Metrics{}
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}

	s.Lock()
	defer s.Unlock()

	rw.Header().Add("Content-Type", "application/json")
	res := s.SetValueInMapJSON(v)

	mt := s.MutexRepo[v.ID].GetMetrics(v.MType, v.ID, s.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := rw.Write(metricsJSON); err != nil {
		constants.Logger.ErrorLog(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res == http.StatusOK {
		var arrMetrics encoding.ArrMetrics
		arrMetrics = append(arrMetrics, mt)

		for _, val := range s.Config.StorageType {
			val.WriteMetric(arrMetrics)
		}
	}
	rw.WriteHeader(res)
}

func (s *serverHTTP) HandlerUpdatesMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	IPAddressAllowed := rq.Context().Value(middlware.KeyValueContext("IP-Address-Allowed"))
	if IPAddressAllowed == "false" {
		return
	}

	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := rq.Header.Get("Content-Encoding")

	bodyJSON = rq.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(rq.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
	}

	var storedData encoding.ArrMetrics
	if err := json.Unmarshal(respByte, &storedData); err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
	}

	s.Lock()
	defer s.Unlock()

	for _, val := range storedData {
		s.SetValueInMapJSON(val)
		s.MutexRepo[val.ID].GetMetrics(val.MType, val.ID, s.Config.Key)
	}

	for _, val := range s.Config.StorageType {
		val.WriteMetric(storedData)
	}
}

func (s *serverHTTP) HandlerValueMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader
	bodyJSON = rq.Body

	fmt.Println("------- HandlerGetValue", 1)

	acceptEncoding := rq.Header.Get("Accept-Encoding")
	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		constants.Logger.InfoLog("-- метрика с агента gzip (value)")
		bytBody, err := io.ReadAll(rq.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.Logger.ErrorLog(err)
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	}

	v := encoding.Metrics{}
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.Logger.ErrorLog(err)
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID
	fmt.Println("------- HandlerGetValue", metType, metName)

	s.Lock()
	defer s.Unlock()

	if _, findKey := s.MutexRepo[metName]; !findKey {

		constants.Logger.InfoLog(fmt.Sprintf("== %d %s %d %s", 1, metName, len(s.MutexRepo), s.Config.DatabaseDsn))

		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := s.MutexRepo[metName].GetMetrics(metType, metName, s.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	var byteMeterics []byte
	bt := bytes.NewBuffer(metricsJSON).Bytes()
	byteMeterics = append(byteMeterics, bt...)
	compData, err := compression.Compress(byteMeterics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	var bodyBate []byte
	rw.Header().Add("Content-Type", "application/json")
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Add("Content-Encoding", "gzip")
		bodyBate = compData
	} else {
		bodyBate = metricsJSON
	}

	if _, err := rw.Write(bodyBate); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
}

func (s *serverHTTP) HandlerPingDB(rw http.ResponseWriter, rq *http.Request) {
	defer rq.Body.Close()
	mapTypeStore := s.Config.StorageType

	if _, findKey := mapTypeStore[constants.MetricsStorageDB.String()]; !findKey {
		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if mapTypeStore[constants.MetricsStorageDB.String()].ConnDB() == nil {

		constants.Logger.ErrorLog(errors.New("соединение с базой отсутствует"))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *serverHTTP) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (s *serverHTTP) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	arrMetricsAndValue := s.MapMetrics.TextMetricsAndValue()

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
	}
	content = content + `</ul>
						</body>
						</html>`

	acceptEncoding := rq.Header.Get("Accept-Encoding")

	metricsHTML := []byte(content)
	byteMterics := bytes.NewBuffer(metricsHTML).Bytes()
	compData, err := compression.Compress(byteMterics)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	var bodyBate []byte
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Add("Content-Encoding", "gzip")
		bodyBate = compData
	} else {
		bodyBate = metricsHTML
	}

	rw.Header().Add("Content-Type", "text/html")
	if _, err := rw.Write(bodyBate); err != nil {
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

func (s *serverHTTP) RestoreData() {
	s.Lock()
	defer s.Unlock()

	for _, val := range s.Config.StorageType {
		arrMetrics, err := val.GetMetric()
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}

		for _, val := range arrMetrics {
			s.SetValueInMapJSON(val)
		}
	}
}

func (s *serverHTTP) BackupData() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	saveTicker := time.NewTicker(s.Config.StoreInterval)
	for {
		select {
		case <-saveTicker.C:

			for _, val := range s.Config.StorageType {
				val.WriteMetric(s.PrepareDataBuckUp())
			}

		case <-ctx.Done():
			cancelFunc()
			return
		}
	}
}

func (s *serverHTTP) HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Метрика "+r.URL.Path+" не найдена", http.StatusNotFound)

}

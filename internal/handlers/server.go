package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/postgresql"
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

func (mt MetricType) String() string {
	return [...]string{"gauge", "counter"}[mt]
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

type RepStore struct {
	Config    environment.ServerConfig
	Router    chi.Router
	MX        sync.Mutex
	MutexRepo repository.MapMetrics
}

func NewRepStore() *RepStore {

	rp := new(RepStore)
	rp.New()

	return rp
}

func (rs *RepStore) New() {

	rs.MutexRepo = make(repository.MapMetrics)
	rs.Router = chi.NewRouter()

	rs.Router.Use(middleware.RequestID)
	rs.Router.Use(middleware.RealIP)
	rs.Router.Use(middleware.Logger)
	rs.Router.Use(middleware.Recoverer)
	rs.Router.Use(middleware.StripSlashes)

	rs.Router.HandleFunc("/", rs.HandleFunc)
	rs.Router.NotFound(HandlerNotFound)

	rs.Router.Get("/", rs.HandlerGetAllMetrics)
	rs.Router.Get("/value/{metType}/{metName}", rs.HandlerGetValue)
	rs.Router.Post("/update/{metType}/{metName}/{metValue}", rs.HandlerSetMetricaPOST)
	rs.Router.Post("/update", rs.HandlerUpdateMetricJSON)
	rs.Router.Post("/updates", rs.HandlerUpdatesMetricJSON)
	rs.Router.Post("/value", rs.HandlerValueMetricaJSON)
	rs.Router.Get("/ping", rs.HandlerPingDB)

	rs.Config = environment.SetConfigServer()
}

func (rs *RepStore) setValueInMap(metType string, metName string, metValue string) int {

	switch metType {
	case GaugeMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valG := repository.Gauge(0)
			if ok := valG.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			rs.MutexRepo[metName] = &valG
		}

	case CounterMetric.String():
		if val, findKey := rs.MutexRepo[metName]; findKey {
			if ok := val.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}
		} else {

			valC := repository.Counter(0)
			if ok := valC.SetFromText(metValue); !ok {
				return http.StatusBadRequest
			}

			rs.MutexRepo[metName] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	return http.StatusOK
}

func (rs *RepStore) SetValueInMapJSON(v encoding.Metrics) int {

	var heshVal string

	switch v.MType {
	case GaugeMetric.String():
		var valValue float64
		valValue = *v.Value

		msg := fmt.Sprintf("%s:gauge:%f", v.ID, valValue)
		heshVal = cryptohash.HeshSHA256(msg, rs.Config.Key)
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valG := repository.Gauge(0)
			rs.MutexRepo[v.ID] = &valG
		}
	case CounterMetric.String():
		var valDelta int64
		valDelta = *v.Delta

		msg := fmt.Sprintf("%s:counter:%d", v.ID, valDelta)
		heshVal = cryptohash.HeshSHA256(msg, rs.Config.Key)
		if _, findKey := rs.MutexRepo[v.ID]; !findKey {
			valC := repository.Counter(0)
			rs.MutexRepo[v.ID] = &valC
		}
	default:
		return http.StatusNotImplemented
	}

	heshAgent := []byte(v.Hash)
	heshServer := []byte(heshVal)

	hmacEqual := hmac.Equal(heshServer, heshAgent)

	constants.InfoLevel.Info().Msgf("-- %s - %s", v.Hash, heshVal)
	if v.Hash != "" && !hmacEqual {
		constants.InfoLevel.Info().Msgf("++ %s - %s", v.Hash, heshVal)
		return http.StatusBadRequest
	}

	constants.InfoLevel.Info().Msgf("** %s %s %v %d", v.ID, v.MType, v.Value, v.Delta)
	rs.MutexRepo[v.ID].Set(v)
	return http.StatusOK

}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	rs.MX.Lock()
	defer rs.MX.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {
		constants.InfoLevel.Info().Msgf("== %d", 3)
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	strMetric := rs.MutexRepo[metName].String()
	_, err := io.WriteString(rw, strMetric)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	rs.MX.Lock()
	defer rs.MX.Unlock()

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	rw.WriteHeader(rs.setValueInMap(metType, metName, metValue))
}

func (rs *RepStore) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader

	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(rq.Body)
		if err != nil {
			constants.InfoLevel.Info().Msgf("$$ 1 %s", err.Error())
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.InfoLevel.Info().Msgf("$$ 2 %s", err.Error())
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	} else {
		bodyJSON = rq.Body
	}

	v := encoding.Metrics{}
	//err := json.NewDecoder(rq.Body).Decode(&v)
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.InfoLevel.Info().Msgf("$$ 3 %s", err.Error())
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}

	rs.MX.Lock()
	defer rs.MX.Unlock()

	//rw.Header().Add("Content-Encoding", "gzip")
	rw.Header().Add("Content-Type", "application/json")
	res := rs.SetValueInMapJSON(v)
	rw.WriteHeader(res)

	mt := rs.MutexRepo[v.ID].GetMetrics(v.MType, v.ID, rs.Config.Key)

	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	if _, err := rw.Write(metricsJSON); err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	//if rs.Config.StoreInterval == time.Duration(0) {
	if res == http.StatusOK {
		rs.SaveMetric(mt)
	}
	//}
}

func (rs *RepStore) HandlerUpdatesMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := rq.Header.Get("Content-Encoding")

	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(rq.Body)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err = compression.Decompress(bytBody)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	} else {
		bodyJSON = rq.Body
	}

	respByte, err := ioutil.ReadAll(bodyJSON)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
	}
	var metricsJSON []encoding.Metrics
	if err := json.Unmarshal(respByte, &metricsJSON); err != nil {
		constants.InfoLevel.Error().Err(err)
		http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
	}

	rs.MX.Lock()
	defer rs.MX.Unlock()

	var sch int64
	for _, val := range metricsJSON {

		rs.SetValueInMapJSON(val)
		rs.MutexRepo[val.ID].GetMetrics(val.MType, val.ID, rs.Config.Key)
		sch++

	}

	//if rs.Config.StoreInterval == time.Duration(0) {
	if sch != 0 {
		var mt encoding.Metrics
		rs.SaveMetric(mt)
		//}
	}
}

func (rs *RepStore) HandlerValueMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader

	acceptEncoding := rq.Header.Get("Accept-Encoding")
	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		constants.InfoLevel.Info().Msg("-- метрика с агента gzip (value)")
		bytBody, err := ioutil.ReadAll(rq.Body)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
			http.Error(rw, "Ошибка получения Content-Encoding", http.StatusInternalServerError)
			return
		}

		arrBody, err := compression.Decompress(bytBody)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
			http.Error(rw, "Ошибка распаковки", http.StatusInternalServerError)
			return
		}

		bodyJSON = bytes.NewReader(arrBody)
	} else {
		bodyJSON = rq.Body
	}

	v := encoding.Metrics{}
	err := json.NewDecoder(bodyJSON).Decode(&v)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		http.Error(rw, "Ошибка получения JSON", http.StatusInternalServerError)
		return
	}
	metType := v.MType
	metName := v.ID

	rs.MX.Lock()
	defer rs.MX.Unlock()

	if _, findKey := rs.MutexRepo[metName]; !findKey {

		constants.InfoLevel.Info().Msgf("== %d %s %d %s", 1, metName, len(rs.MutexRepo), rs.Config.DatabaseDsn)

		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo[metName].GetMetrics(metType, metName, rs.Config.Key)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	var bytMterica []byte
	bt := bytes.NewBuffer(metricsJSON).Bytes()
	bytMterica = append(bytMterica, bt...)
	compData, err := compression.Compress(bytMterica)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
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
		constants.InfoLevel.Error().Err(err)
		return
	}
}

func (rs *RepStore) HandlerPingDB(rw http.ResponseWriter, rq *http.Request) {
	defer rq.Body.Close()

	ctx := context.Background()
	pool, err := postgresql.NewClient(ctx, rs.Config.DatabaseDsn)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
	defer pool.Close(ctx)

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	arrMetricsAndValue := textMetricsAndValue(rs.MutexRepo)

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
		constants.InfoLevel.Error().Err(err)
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
		constants.InfoLevel.Error().Err(err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) SaveMetric(metric encoding.Metrics) {

	if rs.Config.StoreFile == "" && rs.Config.DatabaseDsn == "" {
		return
	}

	var arr []encoding.Metrics
	if metric.ID == "" && metric.MType == "" {
		arr = JSONMetricsAndValue(rs.MutexRepo, rs.Config.Key)
	} else {
		arr = append(arr, metric)
	}

	if rs.Config.StoreFile != "" {
		arrJSON, err := json.Marshal(arr)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
		}
		if err := ioutil.WriteFile(rs.Config.StoreFile, arrJSON, 0777); err != nil {
			constants.InfoLevel.Error().Err(err)
		}
	}

	if rs.Config.DatabaseDsn != "" {
		ctx := context.Background()

		db, err := pgx.Connect(ctx, rs.Config.DatabaseDsn)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
			return
		}
		defer db.Close(ctx)

		tx, err := db.Begin(ctx)
		if err != nil {
			constants.InfoLevel.Error().Err(err)
		}

		for _, val := range arr {

			if err := postgresql.SetMetric2DB(ctx, db, val); err != nil {
				constants.InfoLevel.Error().Err(err)
				constants.InfoLevel.Info().Msgf("@@ %s %s %s %v %d", err.Error(), val.ID, val.MType, val.Value, val.Delta)
				continue
			}

		}

		tx.Commit(ctx)
	}
}

func (rs *RepStore) LoadStoreMetricsDB() {

	ctx := context.Background()
	db, err := postgresql.NewClient(ctx, rs.Config.DatabaseDsn)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
	}
	defer db.Close(ctx)

	arrMatric, err := postgresql.GetMetricFromDB(ctx, db)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	rs.MX.Lock()
	defer rs.MX.Unlock()

	for _, val := range arrMatric {
		rs.SetValueInMapJSON(val)
	}
}

func (rs *RepStore) LoadStoreMetricsFile() {

	res, err := ioutil.ReadFile(rs.Config.StoreFile)
	if err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}
	var arrMatric []encoding.Metrics
	if err := json.Unmarshal(res, &arrMatric); err != nil {
		constants.InfoLevel.Error().Err(err)
		return
	}

	rs.MX.Lock()
	defer rs.MX.Unlock()

	for _, val := range arrMatric {
		rs.SetValueInMapJSON(val)
	}
}

func (rs *RepStore) LoadStoreMetrics() {

	constants.InfoLevel.Info().Msgf("@@ %s - %s", rs.Config.DatabaseDsn, rs.Config.StoreFile)
	if rs.Config.DatabaseDsn != "" {
		constants.InfoLevel.Info().Msg("@@ DB")
		rs.LoadStoreMetricsDB()
	} else {
		constants.InfoLevel.Info().Msg("@@ FILE")
		rs.LoadStoreMetricsFile()
	}
}

func HandlerNotFound(rw http.ResponseWriter, r *http.Request) {

	http.Error(rw, "Метрика "+r.URL.Path+" не найдена", http.StatusNotFound)

}

func textMetricsAndValue(mm repository.MapMetrics) []string {
	const msgFormat = "%s = %s"

	var msg []string

	for key, val := range mm {
		msg = append(msg, fmt.Sprintf(msgFormat, key, val.String()))
	}

	return msg
}

func JSONMetricsAndValue(mm repository.MapMetrics, hashKey string) []encoding.Metrics {

	var arr []encoding.Metrics

	for key, val := range mm {
		jMetric := val.GetMetrics(val.Type(), key, hashKey)
		arr = append(arr, jMetric)
	}

	return arr
}

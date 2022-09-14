package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/postgresql"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type MetricError int

type HTMLParam struct {
	Title       string
	TextMetrics []string
}

func (et MetricError) String() string {
	return [...]string{"Not error", "Error convert", "Error get type"}[et]
}

type RepStore struct {
	Config    environment.ServerConfig
	Router    chi.Router
	MutexRepo repository.StoreMetrics
}

func NewRepStore() *RepStore {

	rp := new(RepStore)
	rp.New()

	return rp
}

func (rs *RepStore) New() {

	//rs.MutexRepo = make(repository.StoreMetrics)
	rs.MutexRepo.Repo = make(repository.MapMetrics)

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
	rs.Router.Post("/value", rs.HandlerValueMetricaJSON)
	rs.Router.Get("/ping", rs.HandlerPingDB)

	dataConfig := new(environment.DataConfig)
	environment.SetConfigServer(dataConfig, rs.Config)

	mapTypeStore := dataConfig.MapTypeStore
	if _, findKey := mapTypeStore[constants.MetricsStorageDB.String()]; findKey {
		ctx := context.Background()
		db, err := postgresql.NewClient(ctx, rs.Config.DatabaseDsn)
		if err != nil {
			constants.Logger.ErrorLog(err)
		}

		mapTypeStore[constants.MetricsStorageDB.String()] = &repository.TypeStoreDataDB{DB: db, Ctx: ctx}
		mapTypeStore[constants.MetricsStorageDB.String()].CreateTable()
	}
	if _, findKey := mapTypeStore[constants.MetricsStorageFile.String()]; findKey {
		mapTypeStore[constants.MetricsStorageFile.String()] = &repository.TypeStoreDataFile{StoreFile: rs.Config.StoreFile}
	}
	rs.MutexRepo.MapTypeStore = mapTypeStore
	rs.MutexRepo.HashKey = dataConfig.HashKey
	rs.MutexRepo.StoreInterval = dataConfig.StoreInterval
}

func (rs *RepStore) HandlerGetValue(rw http.ResponseWriter, rq *http.Request) {

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")

	rs.MutexRepo.MX.Lock()
	defer rs.MutexRepo.MX.Unlock()

	if _, findKey := rs.MutexRepo.Repo[metName]; !findKey {
		constants.Logger.InfoLog(fmt.Sprintf("== %d", 3))
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	strMetric := rs.MutexRepo.Repo[metName].String()
	_, err := io.WriteString(rw, strMetric)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (rs *RepStore) HandlerSetMetricaPOST(rw http.ResponseWriter, rq *http.Request) {

	rs.MutexRepo.MX.Lock()
	defer rs.MutexRepo.MX.Unlock()

	metType := chi.URLParam(rq, "metType")
	metName := chi.URLParam(rq, "metName")
	metValue := chi.URLParam(rq, "metValue")

	rw.WriteHeader(rs.MutexRepo.SetValueInMap(metType, metName, metValue))
}

func (rs *RepStore) HandlerUpdateMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader

	contentEncoding := rq.Header.Get("Content-Encoding")
	bodyJSON = rq.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(rq.Body)
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

	rs.MutexRepo.MX.Lock()
	defer rs.MutexRepo.MX.Unlock()

	rw.Header().Add("Content-Type", "application/json")
	res := rs.MutexRepo.SetValueInMapJSON(v)
	rw.WriteHeader(res)

	mt := rs.MutexRepo.Repo[v.ID].GetMetrics(v.MType, v.ID, rs.MutexRepo.HashKey)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	if _, err := rw.Write(metricsJSON); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	if res == http.StatusOK {
		var arrMetrics encoding.ArrMetrics
		arrMetrics = append(arrMetrics, mt)

		for _, val := range rs.MutexRepo.MapTypeStore {
			val.WriteMetric(arrMetrics)
		}

	}
}

func (rs *RepStore) HandlerUpdatesMetricJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader
	var arrBody []byte

	contentEncoding := rq.Header.Get("Content-Encoding")

	bodyJSON = rq.Body
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(rq.Body)
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

	rs.MutexRepo.MX.Lock()
	defer rs.MutexRepo.MX.Unlock()

	for _, val := range storedData {
		rs.MutexRepo.SetValueInMapJSON(val)
		rs.MutexRepo.Repo[val.ID].GetMetrics(val.MType, val.ID, rs.MutexRepo.HashKey)
	}

	for _, val := range rs.MutexRepo.MapTypeStore {
		val.WriteMetric(storedData)
	}
}

func (rs *RepStore) HandlerValueMetricaJSON(rw http.ResponseWriter, rq *http.Request) {

	var bodyJSON io.Reader
	bodyJSON = rq.Body

	acceptEncoding := rq.Header.Get("Accept-Encoding")
	contentEncoding := rq.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		constants.Logger.InfoLog("-- метрика с агента gzip (value)")
		bytBody, err := ioutil.ReadAll(rq.Body)
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

	rs.MutexRepo.MX.Lock()
	defer rs.MutexRepo.MX.Unlock()

	if _, findKey := rs.MutexRepo.Repo[metName]; !findKey {

		constants.Logger.InfoLog(fmt.Sprintf("== %d %s %d %s", 1, metName, len(rs.MutexRepo.Repo), rs.Config.DatabaseDsn))

		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, "Метрика "+metName+" с типом "+metType+" не найдена", http.StatusNotFound)
		return
	}

	mt := rs.MutexRepo.Repo[metName].GetMetrics(metType, metName, rs.MutexRepo.HashKey)
	metricsJSON, err := mt.MarshalMetrica()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	var bytMterica []byte
	bt := bytes.NewBuffer(metricsJSON).Bytes()
	bytMterica = append(bytMterica, bt...)
	compData, err := compression.Compress(bytMterica)
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

func (rs *RepStore) HandlerPingDB(rw http.ResponseWriter, rq *http.Request) {
	mapTypeStore := rs.MutexRepo.MapTypeStore
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

func (rs *RepStore) HandleFunc(rw http.ResponseWriter, rq *http.Request) {

	defer rq.Body.Close()
	rw.WriteHeader(http.StatusOK)
}

func (rs *RepStore) HandlerGetAllMetrics(rw http.ResponseWriter, rq *http.Request) {

	arrMetricsAndValue := textMetricsAndValue(rs.MutexRepo.Repo)

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

func (rs *RepStore) RestoreData() {
	for _, val := range rs.MutexRepo.MapTypeStore {
		arrMetrics, err := val.GetMetric()
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}

		rs.MutexRepo.MX.Lock()
		defer rs.MutexRepo.MX.Unlock()

		for _, val := range arrMetrics {
			rs.MutexRepo.SetValueInMapJSON(val)
		}
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

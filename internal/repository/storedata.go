package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/jackc/pgx/v4"
	"io/ioutil"
	"os"
)

type TypeStoreDataDB struct {
	DB  *pgx.Conn
	Ctx context.Context
}
type TypeStoreDataFile struct {
	StoreFile string
}

type MapTypeStore = map[string]TypeStoreData

type TypeStoreData interface {
	WriteMetric(storedData encoding.ArrMetrics)
	GetMetric() ([]encoding.Metrics, error)
	CreateTable()
	ConnDB() *pgx.Conn
	SetMetric2DB(storedData encoding.ArrMetrics) error
}

func (sdb *TypeStoreDataDB) WriteMetric(storedData encoding.ArrMetrics) {
	tx, err := sdb.DB.Begin(sdb.Ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	if err := sdb.SetMetric2DB(storedData); err != nil {
		constants.Logger.ErrorLog(err)
	}

	if err := tx.Commit(sdb.Ctx); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (sdb *TypeStoreDataDB) GetMetric() ([]encoding.Metrics, error) {
	var arrMatrics []encoding.Metrics

	poolRow, err := sdb.DB.Query(sdb.Ctx, constants.QuerySelect)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, errors.New("ошибка чтения БД")
	}
	defer poolRow.Close()

	for poolRow.Next() {
		var nst encoding.Metrics

		err = poolRow.Scan(&nst.ID, &nst.MType, &nst.Value, &nst.Delta, &nst.Hash)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}
		arrMatrics = append(arrMatrics, nst)
	}

	return arrMatrics, nil
}

func (sdb *TypeStoreDataDB) ConnDB() *pgx.Conn {
	return sdb.DB
}

func (sdb *TypeStoreDataDB) CreateTable() {

	if _, err := sdb.DB.Exec(sdb.Ctx, constants.QuerySchema); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	if _, err := sdb.DB.Exec(sdb.Ctx, constants.QueryTable); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

func (sdb *TypeStoreDataDB) SetMetric2DB(storedData encoding.ArrMetrics) error {

	for _, data := range storedData {
		rows, err := sdb.DB.Query(sdb.Ctx, constants.QuerySelectWithWhereTemplate, data.ID, data.MType)
		if err != nil {
			return errors.New("ошибка выборки данных в БД")
		}

		dataValue := float64(0)
		if data.Value != nil {
			dataValue = *data.Value
		}
		dataDelta := int64(0)
		if data.Delta != nil {
			dataDelta = *data.Delta
		}

		insert := true
		if rows.Next() {
			insert = false
		}
		rows.Close()

		if insert {
			if _, err := sdb.DB.Exec(sdb.Ctx, constants.QueryInsertTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New(err.Error())
			}
		} else {
			if _, err := sdb.DB.Exec(sdb.Ctx, constants.QueryUpdateTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New("ошибка обновления данных в БД")
			}
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

func (f *TypeStoreDataFile) WriteMetric(storedData encoding.ArrMetrics) {
	arrJSON, err := json.Marshal(storedData)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	if err := ioutil.WriteFile(f.StoreFile, arrJSON, 0777); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
}

func (f *TypeStoreDataFile) GetMetric() ([]encoding.Metrics, error) {
	res, err := ioutil.ReadFile(f.StoreFile)
	if err != nil {
		return nil, err
	}
	var arrMatric []encoding.Metrics
	if err := json.Unmarshal(res, &arrMatric); err != nil {
		return nil, err
	}

	return arrMatric, nil
}

func (f *TypeStoreDataFile) ConnDB() *pgx.Conn {
	return nil
}

func (f *TypeStoreDataFile) CreateTable() {
	if _, err := os.Create(f.StoreFile); err != nil {
		constants.Logger.ErrorLog(err)
	}

}

func (f *TypeStoreDataFile) SetMetric2DB(storedData encoding.ArrMetrics) error {
	for _, val := range storedData {
		constants.Logger.InfoLog(fmt.Sprintf("очень странно, но этого сообщения не должно быть", val.ID))
	}
	return nil
}

package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/postgresql"
)

// TypeStoreDataDB Структура хранения настроек БД
// DBC: конект с базой данных
// Ctx: контекст на момент создания
// DBDsn: строка соединения с базой данных
type TypeStoreDataDB struct {
	DBC   postgresql.DBConnector
	Ctx   context.Context
	DBDsn string
}

// TypeStoreDataFile Структура хранения настроек файла
// StoreFile путь к файлу хранения метрик
type TypeStoreDataFile struct {
	StoreFile string
}

type SyncMapMetrics struct {
	sync.Mutex
	MapMetrics
}

type MapTypeStore = map[string]TypeStoreData

type TypeStoreData interface {
	WriteMetric(storedData encoding.ArrMetrics)
	GetMetric() ([]encoding.Metrics, error)
	CreateTable() bool
	ConnDB() *pgxpool.Pool
}

// InitStoreDB инициализация хранилища БД
func InitStoreDB(mts MapTypeStore, store string) (MapTypeStore, error) {
	if _, findKey := mts[constants.MetricsStorageDB.String()]; findKey {
		ctx := context.Background()

		dbc, err := postgresql.PoolDB(store)
		if err != nil {
			return nil, err
		}

		mts[constants.MetricsStorageDB.String()] = &TypeStoreDataDB{
			DBC: *dbc, Ctx: ctx, DBDsn: store,
		}
		if ok := mts[constants.MetricsStorageDB.String()].CreateTable(); !ok {
			return nil, err
		}
	}
	//if _, findKey := mts[constants.MetricsStorageFile.String()]; findKey {
	//	mts[constants.MetricsStorageDB.String()] = &TypeStoreDataFile{StoreFile: store}
	//}

	return mts, nil
}

// InitStoreFile инициализация хранилища в файле
func InitStoreFile(mts MapTypeStore, store string) (MapTypeStore, error) {

	if _, findKey := mts[constants.MetricsStorageFile.String()]; findKey {
		mts[constants.MetricsStorageDB.String()] = &TypeStoreDataFile{StoreFile: store}
	}

	return mts, nil
}

// WriteMetric Запись метрик в базу данных
func (sdb *TypeStoreDataDB) WriteMetric(storedData encoding.ArrMetrics) {
	dataBase := sdb.DBC
	if err := dataBase.SetMetric2DB(storedData); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

// GetMetric Получение метрик из базы данных
func (sdb *TypeStoreDataDB) GetMetric() ([]encoding.Metrics, error) {
	var arrMatrics []encoding.Metrics

	ctx := context.Background()
	defer ctx.Done()

	conn, err := sdb.DBC.Pool.Acquire(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, errors.New("ошибка создания соединения с БД")
	}
	defer conn.Release()

	poolRow, err := conn.Query(sdb.Ctx, constants.QuerySelect)
	if err != nil {
		conn.Release()
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

	ctx.Done()
	conn.Release()

	return arrMatrics, nil
}

// ConnDB Возвращает соединение с базой данных
func (sdb *TypeStoreDataDB) ConnDB() *pgxpool.Pool {
	fmt.Println("++++++++++++++07", sdb.DBC.Pool)
	return sdb.DBC.Pool
}

// CreateTable Проверка и создание, если таковых нет, таблиц в базе данных
func (sdb *TypeStoreDataDB) CreateTable() bool {
	fmt.Println("++++++++++++++08", sdb.DBC.Pool)
	ctx := context.Background()
	conn, err := sdb.DBC.Pool.Acquire(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return false
	}
	defer conn.Release()
	if _, err := conn.Exec(sdb.Ctx, constants.QuerySchema); err != nil {
		conn.Release()
		constants.Logger.ErrorLog(err)
		return false
	}
	if _, err := conn.Exec(sdb.Ctx, constants.QueryTable); err != nil {
		conn.Release()
		constants.Logger.ErrorLog(err)
		return false
	}
	conn.Release()
	ctx.Done()

	return true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

// WriteMetric Запись метрик в файл
func (f *TypeStoreDataFile) WriteMetric(storedData encoding.ArrMetrics) {
	arrJSON, err := json.Marshal(storedData)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	if err := os.WriteFile(f.StoreFile, arrJSON, 0777); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
}

// GetMetric Получение метрик из файла
func (f *TypeStoreDataFile) GetMetric() ([]encoding.Metrics, error) {
	res, err := os.ReadFile(f.StoreFile)
	if err != nil {
		return nil, err
	}
	var arrMatric []encoding.Metrics
	if err := json.Unmarshal(res, &arrMatric); err != nil {
		return nil, err
	}

	return arrMatric, nil
}

// ConnDB Возвращает с файлом. Для файла не используется. Возвращает nil
func (f *TypeStoreDataFile) ConnDB() *pgxpool.Pool {
	fmt.Println("++++++++++++++04", "file")
	return nil
}

// CreateTable Проверка и создание, если нет, файла для хранения метрик
func (f *TypeStoreDataFile) CreateTable() bool {
	fmt.Println("++++++++++++++05", f.StoreFile)
	if _, err := os.Create(f.StoreFile); err != nil {
		fmt.Println("++++++++++++++06", err)
		constants.Logger.ErrorLog(err)
		return false
	}

	return true
}

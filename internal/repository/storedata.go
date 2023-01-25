package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/postgresql"
)

// StorageDB Структура хранения настроек БД
// DBC: конект с базой данных
// Ctx: контекст на момент создания
// DBDsn: строка соединения с базой данных
type StorageDB struct {
	DBC   postgresql.DBConnector
	DBDsn string
}

// StorageFile Структура хранения настроек файла
// StoreFile путь к файлу хранения метрик
type StorageFile struct {
	StoreFile string
}

type Storage interface {
	WriteMetric(storedData encoding.ArrMetrics)
	GetMetric() ([]encoding.Metrics, error)
	CreateTable() error
	ConnDB() error
}

type SyncMapMetrics struct {
	sync.Mutex
	MapMetrics
}

// NewStorage реализует фабричный метод.
func NewStorage(databaseDsn string, storeFile string) Storage {
	if databaseDsn != "" {
		return newDBStorage(databaseDsn)
	}

	if storeFile != "" {
		return newFileStorage(storeFile)
	}

	return nil
}

func newDBStorage(databaseDsn string) *StorageDB {
	storageDB, err := InitStoreDB(databaseDsn)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil
	}
	return storageDB
}

func newFileStorage(storeFile string) *StorageFile {
	storageFile, err := InitStoreFile(storeFile)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil
	}
	return storageFile
}

// InitStoreDB инициализация хранилища БД
func InitStoreDB(store string) (*StorageDB, error) {

	dbc, err := postgresql.PoolDB(store)
	if err != nil {
		return nil, err
	}

	storageDB := &StorageDB{
		DBC: *dbc, DBDsn: store,
	}
	if err = storageDB.CreateTable(); err != nil {
		return nil, err
	}

	return storageDB, nil
}

// InitStoreFile инициализация хранилища в файле
func InitStoreFile(store string) (*StorageFile, error) {
	return &StorageFile{StoreFile: store}, nil
}

// WriteMetric Запись метрик в базу данных
func (s *StorageDB) WriteMetric(storedData encoding.ArrMetrics) {
	dataBase := s.DBC
	if err := dataBase.SetMetric2DB(storedData); err != nil {
		constants.Logger.ErrorLog(err)
	}
}

// GetMetric Получение метрик из базы данных
func (s *StorageDB) GetMetric() ([]encoding.Metrics, error) {
	var arrMatrics []encoding.Metrics

	ctx := context.Background()
	defer ctx.Done()

	conn, err := s.DBC.Pool.Acquire(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return nil, errors.New("ошибка создания соединения с БД")
	}
	defer conn.Release()

	poolRow, err := conn.Query(ctx, constants.QuerySelect)
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

// ConnDB Возвращает ошибку соедениения с БД
func (s *StorageDB) ConnDB() error {
	if s.DBC.Pool == nil {
		return errors.New("нет соединения с БД")
	}
	return nil
}

// CreateTable Проверка и создание, если таковых нет, таблиц в базе данных
func (s *StorageDB) CreateTable() error {
	ctx := context.Background()
	conn, err := s.DBC.Pool.Acquire(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return err
	}
	defer conn.Release()

	if _, err = conn.Exec(ctx, constants.QuerySchema); err != nil {
		conn.Release()
		constants.Logger.ErrorLog(err)
		return err
	}
	if _, err = conn.Exec(ctx, constants.QueryTable); err != nil {
		conn.Release()
		constants.Logger.ErrorLog(err)
		return err
	}
	conn.Release()
	ctx.Done()

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

// WriteMetric Запись метрик в файл
func (f *StorageFile) WriteMetric(storedData encoding.ArrMetrics) {
	arrJSON, err := json.Marshal(storedData)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	if err := os.WriteFile(f.StoreFile, arrJSON, 0664); err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
}

// GetMetric Получение метрик из файла
func (f *StorageFile) GetMetric() ([]encoding.Metrics, error) {
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

// ConnDB Возвращает ошибку соедениения с файлом
func (f *StorageFile) ConnDB() error {
	return nil
}

// CreateTable Проверка и создание, если нет, файла для хранения метрик
func (f *StorageFile) CreateTable() error {
	if _, err := os.Create(f.StoreFile); err != nil {
		constants.Logger.ErrorLog(err)
		return err
	}

	return nil
}

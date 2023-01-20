package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"

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
	Ctx   context.Context
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
	CreateTable() bool
	ConnDB() *pgxpool.Pool
}

type SyncMapMetrics struct {
	sync.Mutex
	MapMetrics
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

// InitStoreDB инициализация хранилища БД
func InitStoreDB(store string) (*StorageDB, error) {

	ctx := context.Background()

	dbc, err := postgresql.PoolDB(store)
	if err != nil {
		return nil, err
	}

	storageDB := &StorageDB{
		DBC: *dbc, Ctx: ctx, DBDsn: store,
	}
	if ok := storageDB.CreateTable(); !ok {
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

	poolRow, err := conn.Query(s.Ctx, constants.QuerySelect)
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
func (s *StorageDB) ConnDB() *pgxpool.Pool {
	return s.DBC.Pool
}

// CreateTable Проверка и создание, если таковых нет, таблиц в базе данных
func (s *StorageDB) CreateTable() bool {
	ctx := context.Background()
	conn, err := s.DBC.Pool.Acquire(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return false
	}
	defer conn.Release()
	if _, err := conn.Exec(s.Ctx, constants.QuerySchema); err != nil {
		conn.Release()
		constants.Logger.ErrorLog(err)
		return false
	}
	if _, err := conn.Exec(s.Ctx, constants.QueryTable); err != nil {
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
func (f *StorageFile) WriteMetric(storedData encoding.ArrMetrics) {
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

// ConnDB Возвращает с файлом. Для файла не используется. Возвращает nil
func (f *StorageFile) ConnDB() *pgxpool.Pool {
	return nil
}

// CreateTable Проверка и создание, если нет, файла для хранения метрик
func (f *StorageFile) CreateTable() bool {
	if _, err := os.Create(f.StoreFile); err != nil {
		constants.Logger.ErrorLog(err)
		return false
	}

	return true
}

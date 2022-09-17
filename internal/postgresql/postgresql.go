package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DataBase struct {
	DB  pgx.Conn
	Ctx context.Context
}

type Context struct {
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

type DBConnector struct {
	Pool    *pgxpool.Pool
	Context Context
}

func PoolDB(dsn string) (*DBConnector, error) {
	if dsn == "" {
		return new(DBConnector), errors.New("пустой путь к базе")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		fmt.Print(err.Error())
	}
	dbc := DBConnector{
		Pool: pool,
		Context: Context{
			Ctx:        ctx,
			CancelFunc: cancelFunc,
		},
	}
	return &dbc, nil
}

func NewClient(ctx context.Context, dsn string) (*pgx.Conn, error) {
	if dsn == "" {
		return nil, errors.New("пустой путь к базе")
	}

	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (DataBase *DataBase) SetMetric2DB(storedData encoding.ArrMetrics) error {

	for _, data := range storedData {
		rows, err := DataBase.DB.Query(DataBase.Ctx, constants.QuerySelectWithWhereTemplate, data.ID, data.MType)
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
			if _, err := DataBase.DB.Exec(DataBase.Ctx, constants.QueryInsertTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New(err.Error())
			}
		} else {
			if _, err := DataBase.DB.Exec(DataBase.Ctx, constants.QueryUpdateTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New("ошибка обновления данных в БД")
			}
		}
	}
	return nil
}

func (DataBase *DBConnector) SetMetric2DB(storedData encoding.ArrMetrics) error {

	conn, err := DataBase.Pool.Acquire(DataBase.Context.Ctx)
	for _, data := range storedData {
		if err != nil {
			return err
		}
		//defer conn.Release()

		//cnn, _ := sql.Open("sql", "")
		//cnn.QueryContext()
		//rows := DataBase.Pool.QueryRow(DataBase.Context.Ctx, constants.QuerySelectWithWhereTemplate, data.ID, data.MType)
		rows, err := conn.Query(DataBase.Context.Ctx, constants.QuerySelectWithWhereTemplate, data.ID, data.MType)
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
			if _, err := conn.Exec(DataBase.Context.Ctx, constants.QueryInsertTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New(err.Error())
			}
		} else {
			if _, err := conn.Exec(DataBase.Context.Ctx, constants.QueryUpdateTemplate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
				constants.Logger.ErrorLog(err)
				return errors.New("ошибка обновления данных в БД")
			}
		}
	}

	conn.Release()
	return nil
}

package postgresql

import (
	"context"
	"errors"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"

	"github.com/jackc/pgx/v4"
)

type DataBase struct {
	DB  pgx.Conn
	Ctx context.Context
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

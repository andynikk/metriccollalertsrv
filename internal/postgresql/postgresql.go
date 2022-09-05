package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

func NewClient(ctx context.Context, dsn string) (*pgx.Conn, error) {
	if dsn == "" {
		return nil, errors.New("пустой путь к базе")
	}

	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	CreateTable(db)
	return db, nil
}

func SetMetric2DB(ctx context.Context, db *pgx.Conn, data encoding.Metrics) error {

	rows, err := db.Query(ctx, constants.QuerySelectWithWhere, data.ID, data.MType)
	if err != nil {
		//fmt.Println(constants.QuerySelectWithWhere, data.ID, data.MType)
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
		rows.Close()
	}

	if insert {
		if _, err := db.Exec(ctx, constants.QueryInsert, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
			return errors.New("ошибка добавления данных в БД")
		}
	} else {
		if _, err := db.Exec(ctx, constants.QueryUpdate, data.ID, data.MType, dataValue, dataDelta, ""); err != nil {
			return errors.New("ошибка обновления данных в БД")
		}
	}

	return nil
}

func GetMetricFromDB(ctx context.Context, db *pgx.Conn) ([]encoding.Metrics, error) {

	var arrMatrics []encoding.Metrics

	poolRow, err := db.Query(ctx, constants.QuerySelect)
	if err != nil {
		fmt.Println("@@@@@@@@@@@@@@@@@@", 1)
		return nil, errors.New("ошибка чтения БД")
	}
	for poolRow.Next() {
		var nst encoding.Metrics

		err = poolRow.Scan(&nst.ID, &nst.MType, &nst.Value, &nst.Delta, &nst.Hash)
		if err != nil {
			fmt.Println("@@@@@@@@@@@@@@@@@@", 2, &nst.ID, &nst.MType, &nst.Value, &nst.Delta, &nst.Hash)
			continue
			//return nil, errors.New("ошибка получения данных БД")
		}
		arrMatrics = append(arrMatrics, nst)
	}

	return arrMatrics, nil
}

func CreateTable(pool *pgx.Conn) {

	querySchema := `CREATE SCHEMA IF NOT EXISTS metrics`
	if _, err := pool.Exec(context.Background(), querySchema); err != nil {
		fmt.Println(err.Error())
		return
	}

	queryTable := `CREATE TABLE IF NOT EXISTS metrics.store
					(
						"ID" character varying COLLATE pg_catalog."default",
						"MType" character varying COLLATE pg_catalog."default",
						"Value" double precision NOT NULL DEFAULT 0,
						"Delta" integer NOT NULL DEFAULT 0,
						"Hash" character varying COLLATE pg_catalog."default"
					)
					
					TABLESPACE pg_default;
					
					ALTER TABLE IF EXISTS metrics.store
						OWNER to postgres;`

	if _, err := pool.Exec(context.Background(), queryTable); err != nil {
		fmt.Println(err.Error())
	}
}

package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
)

type PostgrePool struct {
	Pool *pgxpool.Pool
	Cfg  environment.ServerConfig
	Ctx  context.Context
}

func NewClient(ctx context.Context, cfg environment.ServerConfig) (*pgxpool.Pool, error) {
	dsn := cfg.DatabaseDsn
	if dsn == "" {
		return nil, errors.New("пустой путь к базе")
	}

	//ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	//defer cancel()

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	CreateTable(pool)
	return pool, nil
}

func SetMetric2DB(pool *pgxpool.Pool, arrDB []encoding.Metrics, data encoding.Metrics) error {

	//querySelect := `SELECT
	//					*
	//				FROM
	//					metrics.store`

	queryInsert := `INSERT INTO 
						metrics.store ("ID", "MType", "Value", "Delta") 
					VALUES
						($1, $2, $3, $4)`

	queryUpdate := `UPDATE 
						metrics.store 
					SET 
						"Value"=$3, "Delta"=$4
					WHERE 
						"ID" = $1 
						and "MType" = $2;`

	ctx := context.Background()
	//
	////poolRow, err := pool.Query(ctx, querySelect, data.ID, data.MType)
	//poolRow, err := pool.Query(ctx, querySelect)
	//
	//var arrDB []encoding.Metrics

	//if err != nil {
	//	return errors.New("ошибка выборки данных в БД")
	//}
	//if poolRow.Next() {
	//	if _, err := pool.Query(
	//		ctx, queryUpdate, data.ID, data.MType, data.Value, data.Delta); err != nil {
	//		return errors.New("ошибка обновления данных в БД")
	//	}
	//} else {
	//	println(3)
	//	if _, err := pool.Query(
	//		ctx, queryInsert, data.ID, data.MType, data.Value, data.Delta); err != nil {
	//		return errors.New("ошибка добавления данных в БД")
	//	}
	//}

	for _, val := range arrDB {
		if val.ID == data.ID && val.MType == data.MType {
			if _, err := pool.Query(ctx, queryUpdate, data.ID, data.MType, data.Value, data.Delta); err != nil {
				return errors.New("ошибка обновления данных в БД")
			}
			return nil
		}
	}

	if _, err := pool.Query(ctx, queryInsert, data.ID, data.MType, data.Value, data.Delta); err != nil {
		return errors.New("ошибка добавления данных в БД")
	}

	return nil
}

func (p *PostgrePool) GetMetricFromDB() ([]encoding.Metrics, error) {

	query := `select * from metrics.store`

	poolRow, err := p.Pool.Query(p.Ctx, query)

	var arrMatrics []encoding.Metrics

	if err != nil {
		return nil, errors.New("ошибка чтения БД")
	}
	for poolRow.Next() {
		nst := encoding.Metrics{}

		err = poolRow.Scan(&nst.ID, &nst.MType, &nst.Value, &nst.Delta, &nst.Hash)
		if err != nil {
			fmt.Println("Ошибка получения записи БД")
			continue
		}
		arrMatrics = append(arrMatrics, nst)
	}

	return arrMatrics, nil
}

func CreateTable(pool *pgxpool.Pool) {

	querySchema := `CREATE SCHEMA IF NOT EXISTS metrics`
	if _, err := pool.Exec(context.Background(), querySchema); err != nil {
		fmt.Println(err.Error())
		return
	}

	queryTable := `CREATE TABLE IF NOT EXISTS metrics.store
				(
					"ID" character varying COLLATE pg_catalog."default",
					"MType" character varying COLLATE pg_catalog."default",
					"Value" double precision,
					"Delta" integer,
					"Hash" character varying COLLATE pg_catalog."default"
				)
			
				TABLESPACE pg_default;
			
				ALTER TABLE IF EXISTS metrics.store
				OWNER to postgres;`

	if _, err := pool.Exec(context.Background(), queryTable); err != nil {
		fmt.Println(err.Error())
	}
}

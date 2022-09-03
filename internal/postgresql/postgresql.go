package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error)
}

type Repositoriy struct {
	client Client
}

func NewClient(ctx context.Context, cfg environment.ServerConfig) (*pgxpool.Pool, error) {
	dsn := cfg.DatabaseDsn
	if dsn == "" {
		return nil, errors.New("пустой путь к базе")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func InsertMetric(ctx context.Context, cfg environment.ServerConfig, metricJSON []encoding.Metrics) error {
	pool, err := NewClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := "INSERT INTO metrics.table_1 (ID, MType, Value, Delta, Hash) VALUES ($1, $2, $3, $4, $5)"
	for _, val := range metricJSON {
		pgxRow := pool.QueryRow(ctx, q, val.ID, val.MType, val.Value, val.Delta, val.Hash)
		fmt.Println(pgxRow)
	}

	return nil
}

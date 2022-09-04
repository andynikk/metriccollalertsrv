package postgresql

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
)

type PostgrePool struct {
	Pool *pgxpool.Pool
	Cfg  environment.ServerConfig
	Ctx  context.Context
	Data []encoding.Metrics
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
	return pool, nil
}

func (p *PostgrePool) InsertMetric() bool {

	query := `INSERT INTO metrics.table_1 ("ID", "MType", "Value", "Delta", "Hash") VALUES ($1, $2, $3, $4, $5) RETURNING "ID"`

	for _, val := range p.Data {

		poolRow := p.Pool.QueryRow(p.Ctx, query, val.ID, val.MType, val.Value, val.Delta, "val.Hash")
		err := poolRow.Scan(&val.ID)
		//pgError, ok := err.(*pgconn.PgError)
		if err != nil {
			return false
		}
	}

	return true
}

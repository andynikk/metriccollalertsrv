package constants

import (
	"os"

	"github.com/rs/zerolog"
)

const (
	AddressServer  = "localhost:8080"
	ReportInterval = 10
	PollInterval   = 2
	StoreInterval  = 300000000000
	StoreFile      = "/tmp/devops-metrics-db.json"
	Restore        = true
	QueryInsert    = `INSERT INTO 
						metrics.store ("ID", "MType", "Value", "Delta", "Hash") 
					VALUES
						($1, $2, $3, $4, $5)`

	QueryUpdate = `UPDATE 
						metrics.store 
					SET 
						"Value"=$3, "Delta"=$4, "Hash"=$5
					WHERE 
						"ID" = $1 
						and "MType" = $2;`

	QuerySelectWithWhere = `SELECT 
						* 
					FROM 
						metrics.store
					WHERE 
						"ID" = $1 
						and "MType" = $2;`

	QuerySelect = `SELECT 
						* 
					FROM 
						metrics.store`
)

var InfoLevel = zerolog.New(os.Stdout).Level(zerolog.InfoLevel)

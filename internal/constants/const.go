package constants

import (
	"github.com/andynikk/metriccollalertsrv/internal/logger"
)

type StorageType int
type TypeServer int

const (
	TypeSrvGRPC TypeServer = iota
	TypeSrvHTTP
)

const (
	MetricsStorageDB StorageType = iota
	MetricsStorageFile

	TimeLivingCertificateYaer   = 10
	TimeLivingCertificateMounth = 0
	TimeLivingCertificateDay    = 0

	AddressServer  = "localhost:8080"
	ReportInterval = 10
	PollInterval   = 2
	StoreInterval  = 300000000000
	StoreFile      = "/tmp/devops-metrics-db.json"
	Restore        = true
	ButchSize      = 10

	TypeEncryption = "sha512"

	QueryInsertTemplate = `INSERT INTO 
						metrics.store ("ID", "MType", "Value", "Delta", "Hash") 
					VALUES
						($1, $2, $3, $4, $5)`

	QueryUpdateTemplate = `UPDATE 
						metrics.store 
					SET 
						"Value"=$3, "Delta"=$4, "Hash"=$5
					WHERE 
						"ID" = $1 
						and "MType" = $2;`

	QuerySelectWithWhereTemplate = `SELECT 
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

	NameDB = `yapracticum`

	QueryCheckExistDB = `SELECT datname FROM pg_database WHERE datname = '%s' ORDER BY 1;`

	QueryDB = `CREATE DATABASE %s`

	QuerySchema = `CREATE SCHEMA IF NOT EXISTS metrics`

	QueryTable = `CREATE TABLE IF NOT EXISTS metrics.store
					(
						"ID" character varying COLLATE pg_catalog."default",
						"MType" character varying COLLATE pg_catalog."default",
						"Value" double precision NOT NULL DEFAULT 0,
						"Delta" bigint NOT NULL DEFAULT 0,
						"Hash" character varying COLLATE pg_catalog."default"
					)
					
					TABLESPACE pg_default;
					
					ALTER TABLE IF EXISTS metrics.store
						OWNER to postgres;`

	SepIPAddress = ";"
)

func (st StorageType) String() string {
	return [...]string{"db", "file"}[st]
}

func (ts TypeServer) String() string {
	return [...]string{"gRPC", "HTTP"}[ts]
}

var Logger logger.Logger
var TrustedSubnet string

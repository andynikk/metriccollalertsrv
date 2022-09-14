package environment

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type AgentConfigENV struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	Key            string        `env:"KEY"`
}

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
	Key            string
}

type ServerConfigENV struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	Key           string        `env:"KEY"`
	DatabaseDsn   string        `env:"DATABASE_DSN"`
}

type ServerConfig struct {
	StoreFile   string
	Restore     bool
	Address     string
	DatabaseDsn string
}

type DataConfig struct {
	ServerConfig  ServerConfig
	MapTypeStore  repository.MapTypeStore
	HashKey       string
	StoreInterval time.Duration
}

func SetConfigAgent() AgentConfig {
	addressPtr := flag.String("a", constants.AddressServer, "имя сервера")
	reportIntervalPtr := flag.Duration("r", constants.ReportInterval*time.Second, "интервал отправки на сервер")
	pollIntervalPtr := flag.Duration("p", constants.PollInterval*time.Second, "интервал сбора метрик")
	keyFlag := flag.String("k", "", "ключ хеширования")
	flag.Parse()

	var cfgENV AgentConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	addressServ := ""
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		addressServ = cfgENV.Address
	} else {
		addressServ = *addressPtr
	}

	var reportIntervalMetric time.Duration
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportIntervalMetric = cfgENV.ReportInterval
	} else {
		reportIntervalMetric = *reportIntervalPtr
	}

	var pollIntervalMetrics time.Duration
	if _, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollIntervalMetrics = cfgENV.PollInterval
	} else {
		pollIntervalMetrics = *pollIntervalPtr
	}

	keyHash := ""
	if _, ok := os.LookupEnv("KEY"); ok {
		keyHash = cfgENV.Key
	} else {
		keyHash = *keyFlag
	}

	return AgentConfig{
		Address:        addressServ,
		ReportInterval: reportIntervalMetric,
		PollInterval:   pollIntervalMetrics,
		Key:            keyHash,
	}

}

func SetConfigServer(tempConfig *DataConfig, serverConfig ServerConfig) {

	addressPtr := flag.String("a", constants.AddressServer, "имя сервера")
	restorePtr := flag.Bool("r", constants.Restore, "восстанавливать значения при старте")
	storeIntervalPtr := flag.Duration("i", constants.StoreInterval, "интервал автосохранения (сек.)")
	storeFilePtr := flag.String("f", constants.StoreFile, "путь к файлу метрик")
	keyFlag := flag.String("k", "", "ключ хеша")
	keyDatabaseDsn := flag.String("d", "", "строка соединения с базой")
	flag.Parse()

	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	addressServ := cfgENV.Address
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		addressServ = *addressPtr
	}

	restoreMetric := cfgENV.Restore
	if _, ok := os.LookupEnv("RESTORE"); !ok {
		restoreMetric = *restorePtr
	}

	storeIntervalMetrics := cfgENV.StoreInterval
	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		storeIntervalMetrics = *storeIntervalPtr
	}

	storeFileMetrics := cfgENV.StoreFile
	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
		storeFileMetrics = *storeFilePtr
	}

	keyHash := cfgENV.Key
	if _, ok := os.LookupEnv("KEY"); !ok {
		keyHash = *keyFlag
	}

	databaseDsn := cfgENV.DatabaseDsn
	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
		databaseDsn = *keyDatabaseDsn
	}

	tempConfig.MapTypeStore = make(repository.MapTypeStore)
	if databaseDsn != "" {
		typeDB := repository.TypeStoreDataDB{}
		tempConfig.MapTypeStore[constants.MetricsStorageDB.String()] = &typeDB
	} else if storeFileMetrics != "" {
		typeFile := repository.TypeStoreDataFile{}
		tempConfig.MapTypeStore[constants.MetricsStorageFile.String()] = &typeFile
	}

	constants.Logger.Log = zerolog.New(os.Stdout).Level(zerolog.InfoLevel)

	serverConfig.StoreFile = storeFileMetrics
	serverConfig.Restore = restoreMetric
	serverConfig.Address = addressServ
	serverConfig.DatabaseDsn = databaseDsn

	tempConfig.HashKey = keyHash
	tempConfig.StoreInterval = storeIntervalMetrics

	//serverConfig.StoreInterval =storeIntervalMetrics
	//serverConfig.StoreFile =storeFileMetrics
	//serverConfig.Restore =restoreMetric
	//serverConfig.Address =addressServ
	//serverConfig.Key =keyHash
	//serverConfig.DatabaseDsn =databaseDsn
}

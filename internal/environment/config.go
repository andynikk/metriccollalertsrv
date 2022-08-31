package environment

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfigENV struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type ServerConfigENV struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
}

type ServerConfig struct {
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Address       string
}

func SetConfigAgent() AgentConfig {
	addressPtr := flag.String("a", "localhost:8080", "имя сервера")
	reportIntervalPtr := flag.Duration("r", 10*time.Second, "интервал отправки на сервер")
	pollIntervalPtr := flag.Duration("p", 2*time.Second, "интервал сбора метрик")
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

	return AgentConfig{
		Address:        addressServ,
		ReportInterval: reportIntervalMetric,
		PollInterval:   pollIntervalMetrics,
	}

}

func SetConfigServer() ServerConfig {

	addressPtr := flag.String("a", "localhost:8080", "имя сервера")
	restorePtr := flag.Bool("r", true, "восстанавливать значения при старте")
	storeIntervalPtr := flag.Duration("i", 300000000000, "интервал автосохранения (сек.)")
	storeFilePtr := flag.String("f", "/tmp/devops-metrics-db.json", "путь к файлу метрик")
	flag.Parse()

	var cfgENV ServerConfigENV
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

	restoreMetric := false
	if _, ok := os.LookupEnv("RESTORE"); ok {
		restoreMetric = cfgENV.Restore
	} else {
		restoreMetric = *restorePtr
	}

	var storeIntervalMetrics time.Duration
	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		storeIntervalMetrics = cfgENV.ReportInterval
	} else {
		storeIntervalMetrics = *storeIntervalPtr
	}

	var storeFileMetrics string
	if _, ok := os.LookupEnv("STORE_FILE"); ok {
		storeFileMetrics = cfgENV.StoreFile
	} else {
		storeFileMetrics = *storeFilePtr
	}

	return ServerConfig{
		StoreInterval: storeIntervalMetrics,
		StoreFile:     storeFileMetrics,
		Restore:       restoreMetric,
		Address:       addressServ,
	}
}

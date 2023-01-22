package environment

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/caarlos0/env/v6"
)

type AgentConfigENV struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	Key            string        `env:"KEY"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	Config         string        `env:"CONFIG"`
	TypeServer     string        `env:"TYPE_SRV"`
}

type AgentConfig struct {
	Address          string
	ReportInterval   time.Duration
	PollInterval     time.Duration
	Key              string
	CryptoKey        string
	ConfigFilePath   string
	IPAddress        string
	StringTypeServer string
}

type AgentConfigFile struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func GetAgentConfigFile(file *string) AgentConfigFile {
	var sConfig AgentConfigFile

	res, err := os.ReadFile(*file)
	if err != nil {
		return sConfig
	}

	out := ParseConfigBytes(res)
	defer out.Reset()

	if err = json.Unmarshal(out.Bytes(), &sConfig); err != nil {
		return sConfig
	}
	if isOSWindows() {
		sConfig.CryptoKey = strings.Replace(sConfig.CryptoKey, "/", "\\", -1)
	}

	return sConfig

}

func InitConfigAgent() *AgentConfig {
	configAgent := AgentConfig{}
	configAgent.InitConfigAgentENV()
	configAgent.InitConfigAgentFlag()
	configAgent.InitConfigAgentFile()
	configAgent.InitConfigAgentDefault()

	return &configAgent
}

func (ac *AgentConfig) InitConfigAgentENV() {

	var cfgENV AgentConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		log.Fatal(err)
	}

	pathFileCfg := ""
	if _, ok := os.LookupEnv("CONFIG"); ok {
		pathFileCfg = cfgENV.Config
	}

	addressServ := ""
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		addressServ = cfgENV.Address
	}

	var reportIntervalMetric time.Duration
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportIntervalMetric = cfgENV.ReportInterval
	}

	var pollIntervalMetrics time.Duration
	if _, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollIntervalMetrics = cfgENV.PollInterval
	}

	keyHash := ""
	if _, ok := os.LookupEnv("KEY"); ok {
		keyHash = cfgENV.Key
	}

	patchCryptoKey := ""
	if _, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		patchCryptoKey = cfgENV.CryptoKey
	}

	typeSrv := ""
	if _, ok := os.LookupEnv("TYPE_SRV"); ok {
		typeSrv = cfgENV.TypeServer
	}

	ac.Address = addressServ
	ac.ReportInterval = reportIntervalMetric
	ac.PollInterval = pollIntervalMetrics
	ac.Key = keyHash
	ac.CryptoKey = patchCryptoKey
	ac.ConfigFilePath = pathFileCfg

	ac.StringTypeServer = typeSrv
}

func (ac *AgentConfig) InitConfigAgentFlag() {

	addressPtr := flag.String("a", "", "имя сервера")
	reportIntervalPtr := flag.Duration("r", 0, "интервал отправки на сервер")
	pollIntervalPtr := flag.Duration("p", 0, "интервал сбора метрик")
	keyFlag := flag.String("k", "", "ключ хеширования")
	cryptoKeyFlag := flag.String("crypto-key", "", "файл с криптоключем")
	fileCfg := flag.String("config", "", "файл с конфигурацией")
	fileCfgC := flag.String("c", "", "файл с конфигурацией")

	flag.Parse()

	pathFileCfg := ""
	if *fileCfg != "" {
		pathFileCfg = *fileCfg
	} else if *fileCfgC != "" {
		pathFileCfg = *fileCfgC
	}

	if ac.Address == "" {
		ac.Address = *addressPtr
	}
	if ac.ReportInterval == 0 {
		ac.ReportInterval = *reportIntervalPtr
	}
	if ac.PollInterval == 0 {
		ac.PollInterval = *pollIntervalPtr
	}
	if ac.Key == "" {
		ac.Key = *keyFlag
	}
	if ac.CryptoKey == "" {
		ac.CryptoKey = *cryptoKeyFlag
	}
	if ac.ConfigFilePath == "" {
		ac.ConfigFilePath = pathFileCfg
	}
}

func (ac *AgentConfig) InitConfigAgentFile() {

	if ac.ConfigFilePath == "" {
		return
	}

	jsonCfg := GetAgentConfigFile(&ac.ConfigFilePath)

	addressServ := jsonCfg.Address
	reportIntervalMetric, _ := time.ParseDuration(jsonCfg.ReportInterval)
	pollIntervalMetrics, _ := time.ParseDuration(jsonCfg.PollInterval)
	patchCryptoKey := jsonCfg.CryptoKey

	if ac.Address == "" {
		ac.Address = addressServ
	}
	if ac.ReportInterval == 0 {
		ac.ReportInterval = reportIntervalMetric
	}
	if ac.PollInterval == 0 {
		ac.PollInterval = pollIntervalMetrics
	}
	if ac.CryptoKey == "" {
		ac.CryptoKey = patchCryptoKey
	}
}

func (ac *AgentConfig) InitConfigAgentDefault() {

	addressServ := constants.AddressServer
	reportIntervalMetric := constants.ReportInterval * time.Second
	pollIntervalMetrics := constants.PollInterval * time.Second
	typeSrv := constants.TypeSrvHTTP.String()

	if ac.Address == "" {
		ac.Address = addressServ
	}
	if ac.ReportInterval == 0 {
		ac.ReportInterval = reportIntervalMetric
	}
	if ac.PollInterval == 0 {
		ac.PollInterval = pollIntervalMetrics
	}
	if ac.StringTypeServer == "" {
		ac.StringTypeServer = typeSrv
	}

	ipLocal, err := getLocalIPAddress(addressServ)
	if err != nil {
		ac.IPAddress = constants.IPAddress
	}
	ac.IPAddress = ipLocal
}

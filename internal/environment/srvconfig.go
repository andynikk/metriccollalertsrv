package environment

import (
	"encoding/json"
	"flag"
	"net"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
)

type ServerConfigENV struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	Key           string        `env:"KEY"`
	DatabaseDsn   string        `env:"DATABASE_DSN"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
	Config        string        `env:"CONFIG"`
	TrustedSubnet string        `env:"TRUSTED_SUBNET"`
	TypeServer    string        `env:"TYPE_SRV"`
}

type ServerConfig struct {
	StoreInterval  time.Duration
	StoreFile      string
	Restore        bool
	Address        string
	Key            string
	DatabaseDsn    string
	StorageType    repository.MapTypeStore
	CryptoKey      string
	ConfigFilePath string
	TypeServer     string
	TrustedSubnet  *net.IPNet
}

type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

func GetServerConfigFile(file *string) ServerConfigFile {
	var sConfig ServerConfigFile

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
		sConfig.StoreFile = strings.Replace(sConfig.StoreFile, "/", "\\", -1)
	}

	return sConfig

}

func InitConfigServer() *ServerConfig {
	constants.Logger.Log = zerolog.New(os.Stdout).Level(zerolog.InfoLevel)

	sc := ServerConfig{}
	sc.InitConfigServerENV()
	sc.InitConfigServerFlag()
	sc.InitConfigServerFile()
	sc.InitConfigServerDefault()

	sc.StorageType, _ = repository.InitStoreDB(sc.StorageType, sc.DatabaseDsn)
	sc.StorageType, _ = repository.InitStoreFile(sc.StorageType, sc.StoreFile)

	return &sc
}

func (sc *ServerConfig) InitConfigServerENV() {

	var cfgENV ServerConfigENV
	err := env.Parse(&cfgENV)
	if err != nil {
		return
	}

	var addressServ string
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		addressServ = cfgENV.Address
	}

	var restoreMetric bool
	if _, ok := os.LookupEnv("RESTORE"); ok {
		restoreMetric = cfgENV.Restore
	}

	var storeIntervalMetrics time.Duration
	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		storeIntervalMetrics = cfgENV.StoreInterval
	}

	var storeFileMetrics string
	if _, ok := os.LookupEnv("STORE_FILE"); ok {
		storeFileMetrics = cfgENV.StoreFile
	}

	keyHash := cfgENV.Key
	if _, ok := os.LookupEnv("KEY"); !ok {
		keyHash = cfgENV.Key
	}

	var databaseDsn string
	if _, ok := os.LookupEnv("DATABASE_DSN"); ok {
		databaseDsn = cfgENV.DatabaseDsn
	}

	var patchCryptoKey string
	if _, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		patchCryptoKey = cfgENV.CryptoKey
	}

	var patchFileConfig string
	if _, ok := os.LookupEnv("CONFIG"); ok {
		patchFileConfig = cfgENV.Config
	}

	var trustedSubnet string
	if _, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		trustedSubnet = cfgENV.TrustedSubnet
	}

	var typeSrv string
	if _, ok := os.LookupEnv("TYPE_SRV"); ok {
		typeSrv = cfgENV.TypeServer
	}

	MapTypeStore := make(repository.MapTypeStore)
	if databaseDsn != "" {
		typeDB := repository.TypeStoreDataDB{}
		MapTypeStore[constants.MetricsStorageDB.String()] = &typeDB
	} else if storeFileMetrics != "" {
		typeFile := repository.TypeStoreDataFile{}
		MapTypeStore[constants.MetricsStorageFile.String()] = &typeFile
	}

	sc.StoreInterval = storeIntervalMetrics
	sc.StoreFile = storeFileMetrics
	sc.Restore = restoreMetric
	sc.Address = addressServ
	sc.Key = keyHash
	sc.DatabaseDsn = databaseDsn
	sc.StorageType = MapTypeStore
	sc.CryptoKey = patchCryptoKey
	sc.ConfigFilePath = patchFileConfig
	sc.TypeServer = typeSrv

	_, ipv4Net, _ := net.ParseCIDR(trustedSubnet)
	sc.TrustedSubnet = ipv4Net

	constants.TrustedSubnet = ipv4Net.String()

}

func (sc *ServerConfig) InitConfigServerFlag() {

	addressPtr := flag.String("a", "", "имя сервера")
	restorePtr := flag.Bool("r", false, "восстанавливать значения при старте")
	storeIntervalPtr := flag.Duration("i", 0, "интервал автосохранения (сек.)")
	storeFilePtr := flag.String("f", "", "путь к файлу метрик")
	keyFlag := flag.String("k", "", "ключ хеша")
	keyDatabaseDsn := flag.String("d", "", "строка соединения с базой")
	cryptoKeyFlag := flag.String("crypto-key", "", "файл с криптоключем")
	fileCfg := flag.String("config", "", "файл с конфигурацией")
	fileCfgC := flag.String("c", "", "файл с конфигурацией")
	trustedSubnet := flag.String("t", "", "строковое представление бесклассовой адресации (CIDR)")

	flag.Parse()

	pathFileCfg := ""
	if *fileCfg != "" {
		pathFileCfg = *fileCfg
	} else if *fileCfgC != "" {
		pathFileCfg = *fileCfgC
	}

	if sc.Address == "" {
		sc.Address = *addressPtr
	}
	if sc.StoreInterval == 0 {
		sc.StoreInterval = *storeIntervalPtr
	}
	if sc.StoreFile == "" {
		sc.StoreFile = *storeFilePtr
	}
	if !sc.Restore {
		sc.Restore = *restorePtr
	}
	if sc.Key == "" {
		sc.Key = *keyFlag
	}
	if sc.DatabaseDsn == "" {
		sc.DatabaseDsn = *keyDatabaseDsn
	}
	if sc.CryptoKey == "" {
		sc.CryptoKey = *cryptoKeyFlag
	}
	if sc.ConfigFilePath == "" {
		sc.ConfigFilePath = pathFileCfg
	}
	if len(sc.StorageType) == 0 {

		MapTypeStore := make(repository.MapTypeStore)
		if len(sc.StorageType) == 0 {
			if *keyDatabaseDsn != "" {
				typeDB := repository.TypeStoreDataDB{}
				MapTypeStore[constants.MetricsStorageDB.String()] = &typeDB
			} else if *cryptoKeyFlag != "" {
				typeFile := repository.TypeStoreDataFile{}
				MapTypeStore[constants.MetricsStorageFile.String()] = &typeFile
			}
		}

		sc.StorageType = MapTypeStore
	}
	if sc.TrustedSubnet == nil {
		_, ipv4Net, _ := net.ParseCIDR(*trustedSubnet)
		sc.TrustedSubnet = ipv4Net

		constants.TrustedSubnet = ipv4Net.String()
	}
}

func (sc *ServerConfig) InitConfigServerFile() {

	if sc.ConfigFilePath == "" {
		return
	}

	jsonCfg := GetServerConfigFile(&sc.ConfigFilePath)

	addressServ := jsonCfg.Address
	restoreMetric := jsonCfg.Restore
	storeIntervalMetrics, _ := time.ParseDuration(jsonCfg.StoreInterval)
	storeFileMetrics := jsonCfg.StoreFile
	databaseDsn := jsonCfg.DatabaseDsn
	patchCryptoKey := jsonCfg.CryptoKey
	trustedSubnet := jsonCfg.TrustedSubnet

	if sc.Address == "" {
		sc.Address = addressServ
	}
	if sc.StoreInterval == 0 {
		sc.StoreInterval = storeIntervalMetrics
	}
	if sc.StoreFile == "" {
		sc.StoreFile = storeFileMetrics
	}
	if !sc.Restore {
		sc.Restore = restoreMetric
	}
	if sc.DatabaseDsn == "" {
		sc.DatabaseDsn = databaseDsn
	}
	if sc.CryptoKey == "" {
		sc.CryptoKey = patchCryptoKey
	}

	if len(sc.StorageType) == 0 {
		MapTypeStore := make(repository.MapTypeStore)
		if len(sc.StorageType) == 0 {
			if databaseDsn != "" {
				typeDB := repository.TypeStoreDataDB{}
				MapTypeStore[constants.MetricsStorageDB.String()] = &typeDB
			} else if storeFileMetrics != "" {
				typeFile := repository.TypeStoreDataFile{}
				MapTypeStore[constants.MetricsStorageFile.String()] = &typeFile
			}
		}
		sc.StorageType = MapTypeStore
	}
	if sc.TrustedSubnet == nil {
		_, ipv4Net, _ := net.ParseCIDR(trustedSubnet)
		sc.TrustedSubnet = ipv4Net

		constants.TrustedSubnet = ipv4Net.String()
	}
}

func (sc *ServerConfig) InitConfigServerDefault() {

	if sc.Address == "" {
		sc.Address = constants.AddressServer
	}
	if sc.StoreInterval == 0 {
		sc.StoreInterval = constants.StoreInterval
	}
	if sc.StoreFile == "" {
		sc.StoreFile = constants.StoreFile
	}
	if !sc.Restore {
		sc.Restore = constants.Restore
	}

	if sc.TypeServer == "" {
		sc.TypeServer = constants.TypeSrvHTTP.String()
		//sc.TypeServer = constants.TypeSrvGRPC.String()
	}
}

package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Address of http server.
	Address string `json:"address,omitempty"`

	// StoreFile - полное имя файла, куда сохраняются текущие значения.
	StoreFile string `json:"store_file,omitempty"`

	DatabaseDSN string `json:"database_dsn,omitempty"`

	// RSAPrivateKey stores RSA private key.
	RSAPrivateKey string `json:"crypto_key"`

	SignKey string

	// StoreInterval - интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
	StoreInterval uint64 `json:"store_interval,omitempty"`

	// Restore - загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
	Restore bool `json:"restore,omitempty"`
}

// Merge добавляет параметры из donor если они не пустые.
func (ths *Config) Merge(donor Config) {

	// default = true
	ths.Restore = donor.Restore

	// string
	if donor.Address != "" {
		ths.Address = donor.Address
	}
	if donor.StoreFile != "" {
		ths.StoreFile = donor.StoreFile
	}
	if donor.DatabaseDSN != "" {
		ths.DatabaseDSN = donor.DatabaseDSN
	}
	if donor.RSAPrivateKey != "" {
		ths.RSAPrivateKey = donor.RSAPrivateKey
	}
	if donor.SignKey != "" {
		ths.SignKey = donor.SignKey
	}
	// int
	if donor.StoreInterval != 0 {
		ths.StoreInterval = donor.StoreInterval
	}
}

func FromFlags() FlagsConfig {
	cfg := FlagsConfig{}
	flag.StringVar(&cfg.Address, "a", "", "Net address host:port")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.RSAPrivateKey, "crypto-key", "", "Private key path")
	// file storage
	flag.Uint64Var(&cfg.StoreInterval, "i", 300, "Store interval. Sync on 0")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/metrics-db.json", "Metrics file storage")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore metrics from file storage")

	flag.StringVar(&cfg.SignKey, "k", "", "Hash key")

	var configPathShort string
	flag.StringVar(&configPathShort, "c", "", "Config path short alias")
	flag.StringVar(&cfg.ConfigPath, "config", "", "Config path")

	flag.Parse()

	if cfg.ConfigPath == "" && configPathShort != "" {
		cfg.ConfigPath = configPathShort
	}

	return cfg
}

type EnvConfig struct {
	// ENV[CONFIG]
	ConfigPath string

	Config
}

func FromEnv() EnvConfig {
	c := EnvConfig{
		Config: Config{
			Restore: true,
		},
	}

	if envVal, exists := os.LookupEnv("ADDRESS"); exists {
		c.Address = envVal
	}
	if envVal, exists := os.LookupEnv("KEY"); exists {
		c.SignKey = envVal
	}
	if envVal, exists := os.LookupEnv("DATABASE_DSN"); exists {
		c.DatabaseDSN = envVal
	}
	if envVal, exists := os.LookupEnv("CRYPTO_KEY"); exists {
		c.RSAPrivateKey = envVal
	}

	if envVal, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		storeInterval, err := strconv.Atoi(envVal)
		if err != nil {
			log.Fatal(err)
		}
		if storeInterval < 0 {
			log.Fatal("STORE_INTERVAL must be greater or equal 0")
		}
		c.StoreInterval = uint64(storeInterval)
	}

	if envVal, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		envVal = strings.TrimSpace(envVal)
		if envVal == "" {
			log.Fatal("FILE_STORAGE_PATH must not be empty")
		}
		c.StoreFile = envVal
	}

	if envVal, exists := os.LookupEnv("RESTORE"); exists {
		restore, err := strconv.ParseBool(envVal)
		if err != nil {
			log.Fatal(err)
		}
		c.Restore = restore
	}

	if envVal, exists := os.LookupEnv("CONFIG"); exists {
		c.ConfigPath = envVal
	}

	return c
}

type FlagsConfig struct {
	// -c / -config
	ConfigPath string

	Config
}

func FromJSONFile(path string) Config {
	cfg := Config{
		Restore: true,
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	if err2 := decoder.Decode(&cfg); err2 != nil {
		log.Fatal(err)
	}

	return cfg
}

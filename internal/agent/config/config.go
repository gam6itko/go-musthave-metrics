package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	// аналог переменной окружения ADDRESS или флага -a
	Address string `json:"address"`

	// аналог переменной окружения CRYPTO_KEY или флага -crypto-key
	RSAPublicKey string `json:"crypto_key"`

	SignKey string

	// аналог переменной окружения REPORT_INTERVAL или флага -r
	ReportInterval uint64 `json:"report_interval"`
	// аналог переменной окружения POLL_INTERVAL или флага -p
	PollInterval uint64 `json:"poll_interval"`

	RateLimit uint64

	// Имитация определенного клиентского IP
	XRealIP string `json:"x-real-ip"`
}

// Merge добавляет параметры из donor если они не пустые.
func (ths *Config) Merge(donor Config) {
	// string
	if donor.Address != "" {
		ths.Address = donor.Address
	}
	if donor.RSAPublicKey != "" {
		ths.RSAPublicKey = donor.RSAPublicKey
	}
	if donor.SignKey != "" {
		ths.SignKey = donor.SignKey
	}
	if donor.XRealIP != "" {
		ths.XRealIP = donor.XRealIP
	}
	// int
	if donor.ReportInterval != 0 {
		ths.ReportInterval = donor.ReportInterval
	}
	if donor.PollInterval != 0 {
		ths.PollInterval = donor.PollInterval
	}
}

type FlagsConfig struct {

	// -c / -config
	ConfigPath string

	Config
}

func FromFlags() FlagsConfig {
	cfg := FlagsConfig{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Server address host:port")
	flag.Uint64Var(&cfg.ReportInterval, "r", 10, "Report interval")
	flag.Uint64Var(&cfg.PollInterval, "p", 2, "Poll interval")
	flag.Uint64Var(&cfg.RateLimit, "l", 0, "Request rate limit")
	flag.StringVar(&cfg.RSAPublicKey, "crypto-key", "", "Public key")

	flag.StringVar(&cfg.SignKey, "k", "", "Hash key")
	flag.StringVar(&cfg.XRealIP, "ip", "", "Set X-Real-IP header")

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
	c := EnvConfig{}

	if envVal, exists := os.LookupEnv("ADDRESS"); exists {
		c.Address = envVal
	}
	if envVal, exists := os.LookupEnv("KEY"); exists {
		c.SignKey = envVal
	}

	// uint
	if envVal, exists := os.LookupEnv("POLL_INTERVAL"); exists {
		val, err := strconv.ParseUint(envVal, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		c.PollInterval = val
	}
	if envVal, exists := os.LookupEnv("REPORT_INTERVAL"); exists {
		val, err := strconv.ParseUint(envVal, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		c.ReportInterval = val
	}

	if envVal := os.Getenv("RATE_LIMIT"); envVal != "" {
		val, err := strconv.ParseUint(envVal, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		c.RateLimit = val
	}

	if envVal, exists := os.LookupEnv("CRYPTO_KEY"); exists {
		c.RSAPublicKey = envVal
	}

	if envVal, exists := os.LookupEnv("CONFIG"); exists {
		c.ConfigPath = envVal
	}

	return c
}

func FromJSONFile(path string) Config {
	cfg := Config{}

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

package file

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
	StoreInterval uint64
	// полное имя файла, куда сохраняются текущие значения
	FileStoragePath string
	// загружать или нет ранее сохранённые значения из указанного файла при старте сервера
	Restore bool
}

func FromFlags(c *Config, flagSet *flag.FlagSet) {
	flagSet.Uint64Var(&c.StoreInterval, "i", 300, "Store interval. Sync on 0")
	flagSet.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "Metrics file storage")
	flagSet.BoolVar(&c.Restore, "r", true, "Restore metrics from file storage")
}

func FromEnv(c *Config) error {
	if envVal, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		storeInterval, err := strconv.Atoi(envVal)
		if err != nil {
			return err
		}
		if storeInterval < 0 {
			return errors.New("STORE_INTERVAL must be greater or equal 0")
		}
		c.StoreInterval = uint64(storeInterval)
	}

	if filePath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		filePath = strings.Trim(filePath, " \n\t")
		if filePath == "" {
			return errors.New("FILE_STORAGE_PATH must not be empty")
		}
		c.FileStoragePath = filePath
	}

	if envVal, exists := os.LookupEnv("RESTORE"); exists {
		restore, err := strconv.ParseBool(envVal)
		if err != nil {
			return err
		}
		c.Restore = restore
	}

	return nil
}

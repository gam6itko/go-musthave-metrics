package main

import (
	serverConfig "github.com/gam6itko/go-musthave-metrics/internal/server/config"
)

// initConfig собирает конфиг из разных мест и мержит его по приоритетам.
func initConfig() serverConfig.Config {
	cfg := serverConfig.Config{
		Address:       "localhost:8080",
		StoreFile:     "/tmp/metrics-db.json",
		StoreInterval: 300,
		Restore:       true,
	}

	cfgFlag := serverConfig.FromFlags()
	cfgEnv := serverConfig.FromEnv()

	// from file
	if cfgFlag.ConfigPath != "" || cfgEnv.ConfigPath != "" {
		path := cfgEnv.ConfigPath
		if cfgFlag.ConfigPath != "" {
			path = cfgFlag.ConfigPath
		}
		cfg.Merge(serverConfig.FromJSONFile(path))
	}

	cfg.Merge(cfgEnv.Config)
	cfg.Merge(cfgFlag.Config)

	return cfg
}

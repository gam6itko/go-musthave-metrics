package main

import agentConfig "github.com/gam6itko/go-musthave-metrics/internal/agent/config"

// initConfig собирает конфиг из разных мест и мержит его по приоритетам.
func initConfig() agentConfig.Config {
	cfg := agentConfig.Config{
		Address: "localhost:8080",
	}

	cfgFlag := agentConfig.FromFlags()
	cfgEnv := agentConfig.FromEnv()

	// from file
	if cfgFlag.ConfigPath != "" || cfgEnv.ConfigPath != "" {
		path := cfgEnv.ConfigPath
		if cfgFlag.ConfigPath != "" {
			path = cfgFlag.ConfigPath
		}
		cfg.Merge(agentConfig.FromJSONFile(path))
	}

	cfg.Merge(cfgEnv.Config)
	cfg.Merge(cfgFlag.Config)

	return cfg
}

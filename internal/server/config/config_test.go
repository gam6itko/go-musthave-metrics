package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Merge(t *testing.T) {
	cfg := &Config{}
	donor := Config{
		Address:       "Address",
		StoreInterval: 777,
		StoreFile:     "StoreFile",
		Restore:       true,
		DatabaseDSN:   "DatabaseDSN",
		RSAPrivateKey: "RSAPrivateKey",
		SignKey:       "SignKey",
	}

	cfg.Merge(donor)
	require.Equal(t, "Address", cfg.Address)
	require.Equal(t, uint64(777), cfg.StoreInterval)
	require.Equal(t, "StoreFile", cfg.StoreFile)
	require.True(t, cfg.Restore)
	require.Equal(t, "DatabaseDSN", cfg.DatabaseDSN)
	require.Equal(t, "RSAPrivateKey", cfg.RSAPrivateKey)
	require.Equal(t, "SignKey", cfg.SignKey)
}

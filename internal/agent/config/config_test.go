package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Merge(t *testing.T) {
	cfg := &Config{}
	donor := Config{
		Address:        "Address",
		ReportInterval: 777,
		PollInterval:   888,
		RSAPublicKey:   "RSAPublicKey",
		SignKey:        "SignKey",
	}

	cfg.Merge(donor)
	require.Equal(t, "Address", cfg.Address)
	require.Equal(t, uint64(777), cfg.ReportInterval)
	require.Equal(t, uint64(888), cfg.PollInterval)
	require.Equal(t, "RSAPublicKey", cfg.RSAPublicKey)
	require.Equal(t, "SignKey", cfg.SignKey)
}

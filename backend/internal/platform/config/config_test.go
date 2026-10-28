package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("GIN_MODE", "")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "release", cfg.GinMode)
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("GIN_MODE", "debug")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "debug", cfg.GinMode)
}

func TestLoad_InvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "70000"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("PORT", port)

			_, err := Load()

			require.Error(t, err)
		})
	}
}

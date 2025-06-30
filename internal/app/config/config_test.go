//go:build unit

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Load config successfully", func(t *testing.T) {
		config := MustInitConfig("../../../.env.sample")

		assert.Equal(t, LogLeveler("info"), config.LogLevel)
		assert.Equal(t, false, config.TracingEnabled)
		assert.Equal(t, 3001, config.HTTP.Port)
		assert.Equal(t, false, config.HTTP.PprofEnabled)
		assert.Equal(t, 3002, config.HTTP.PprofPort)
		assert.Equal(t, "postgres://docker@postgres/transaction_development?sslmode=disable", config.DB.DSN)
		assert.Equal(t, 2, config.DB.MaxOpenConnections)
		assert.Equal(t, 1, config.DB.MaxIdleConnections)
		assert.Equal(t, 1*time.Hour, config.DB.MaxConnectionLifetime)
	})
}

func TestDefaultValues(t *testing.T) {
	config := MustInitConfig("../../../test/api/fixtures/.env.dummy")
	assert.Equal(t, LogLeveler("info"), config.LogLevel)
}

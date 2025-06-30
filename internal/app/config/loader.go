package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// MustInitConfig initializes configuration from .env file and returns config structure.
func MustInitConfig(configFile string) Config {
	var (
		vpr = viper.New()
		cfg Config
	)

	vpr.SetConfigFile(configFile)
	vpr.SetConfigType("env")
	vpr.AutomaticEnv()

	// default values
	vpr.SetDefault("LOG_LEVEL", "info")

	if err := vpr.ReadInConfig(); err != nil {
		slog.Error("cannot read local config file", slog.String("error", err.Error()))

		panic(err)
	}

	if err := vpr.Unmarshal(&cfg); err != nil {
		slog.Error("cannot unmarshal config file", slog.String("error", err.Error()))

		panic(err)
	}

	vpr.WatchConfig()

	slog.Debug("cfg", slog.String("config", fmt.Sprintf("%v", cfg)))

	return cfg
}

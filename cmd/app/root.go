package app

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "app",
		Short: "Go Event Source",
	}
	cfgFilePath string
)

func init() { //nolint:gochecknoinits
	rootCmd.PersistentFlags().StringVarP(&cfgFilePath, "config", "c", ".env", "")
	rootCmd.AddCommand(
		httpServerCmd,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("error executing root command", slog.String("error", err.Error()))
	}
}

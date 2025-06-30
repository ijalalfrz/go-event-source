package db

import (
	"database/sql"
	"log/slog"

	"github.com/ijalalfrz/go-event-source/internal/app/config"
	// Register PostgreSQL driver.
	_ "github.com/lib/pq"
)

// InitDB initialises PG database connection.
func InitDB(cfg config.Config) *sql.DB {
	slog.Debug("connecting to DB...", slog.String("dsn", cfg.DB.DSN))

	database, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		slog.Error("could not connect to DB", slog.String("error", err.Error()))

		panic(err)
	}

	// Sets the maximum number of open connections to the database.
	database.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	// Sets the maximum number of connections in the idle connection pool.
	database.SetMaxIdleConns(cfg.DB.MaxIdleConnections)
	// Sets the maximum amount of time a connection may be reused.
	database.SetConnMaxLifetime(cfg.DB.MaxConnectionLifetime)
	// Sets the maximum amount of time a connection may be idle.
	database.SetConnMaxIdleTime(cfg.DB.MaxIdleConnectionTime)

	return database
}

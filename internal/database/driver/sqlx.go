package driver

import (
	"banner/pkg/lib/sl"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type SQLXConfig struct {
	DriverName     string
	DataSourceName string
	MaxOpenConns   int
	MaxIdleConns   int
	MaxLifetime    time.Duration
}

func (c *SQLXConfig) NewSQLXDatabase(log *slog.Logger) (*sqlx.DB, error) {
	const op = "database.driver.sqlx.NewSQLXDatabase"

	log = log.With(
		slog.String("op", op),
	)

	db, err := sqlx.Open(c.DriverName, c.DataSourceName)
	if err != nil {
		log.Error("failed to open database", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info(
		"database parameters",
		slog.Int("max number of open connections", c.MaxOpenConns),
		slog.Int("max number of idle connections", c.MaxIdleConns),
		slog.Duration("max lifetime of open connection", c.MaxLifetime),
	)

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.MaxLifetime)

	if err = db.Ping(); err != nil {
		log.Error("failed to ping database", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}

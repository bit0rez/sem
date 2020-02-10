package config

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Main struct {
	DebugMode bool
	DbDriver  string
	DbPath    string
	LogLevel  uint
}

func (m *Main) Validate() error {
	if m.LogLevel > 6 {
		return errors.New("Log level MUST be one of [0..6]")
	}

	if m.DbPath == "" {
		return errors.New("Database path MUST be specified")
	}

	if !validSQLDriver(m.DbDriver) {
		return fmt.Errorf("Database driver MUST be one of %+v", sql.Drivers())
	}

	return nil
}

func ParseFlags() (*Main, error) {
	c := &Main{}

	// Parse flags
	flag.BoolVar(&c.DebugMode, "debug", false, "Enable debug mode")
	flag.UintVar(&c.LogLevel, "loglevel", uint(log.InfoLevel), "Log level. Variants: 0..6.")
	flag.StringVar(
		&c.DbDriver,
		"dbdriver",
		"sqlite3",
		fmt.Sprintf("Database driver. Variants: %v", sql.Drivers()),
	)
	flag.StringVar(&c.DbPath, "dbpath", "", "Database path. Required.")

	flag.Parse()

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func validSQLDriver(driver string) bool {
	for _, d := range sql.Drivers() {
		if d == driver {
			return true
		}
	}
	return false
}

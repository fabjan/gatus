package storage

import (
	"errors"
	"time"
)

// Config is the configuration for alerting providers
type Config struct {
	// File is the path of the file to use for persistence
	// If blank, file persistence is disabled.
	File string `yaml:"file"`
	// PostgresTable is the name of the database table to use for persistence
	// If blank, database persistence is disabled.
	PostgresTable string `yaml:"postgres-table"`
	// AutoSaveInterval is the interval between persisting status information
	AutoSaveInterval time.Duration `yaml:"auto-save-interval"`
}

// ValidateAndSetDefaults checks and sets the default values for fields that are not set
func (cfg *Config) ValidateAndSetDefaults() error {

	// validate
	if cfg.AutoSaveInterval < 0 {
		return errors.New("invalid auto save interval: value should be greater than 0")
	}

	// and set defaults
	if cfg.AutoSaveInterval == time.Duration(0) {
		cfg.AutoSaveInterval = 7 * time.Minute
	}

	return nil
}

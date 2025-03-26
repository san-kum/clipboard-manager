package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	StoragePath string
	MaxEntries  int
	LogLevel    logrus.Level
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		StoragePath: filepath.Join(homeDir, ".clipman", "history"),
		MaxEntries:  1000,
		LogLevel:    logrus.InfoLevel,
	}
}

func (c *Config) Validate() error {
	if err := os.MkdirAll(c.StoragePath, 0755); err != nil {
		return errors.Wrap(err, "failed to create storage directory")
	}
	if c.MaxEntries <= 0 {
		c.MaxEntries = 1000
	}
	return nil
}

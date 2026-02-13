package config

import (
	"errors"
	"fmt"
	"strings"

	wbfconf "github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Migrations MigrationsConfig `yaml:"migrations"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type ServerConfig struct {
	Addr               string `yaml:"addr"`
	ShutdownTimeoutSec int    `yaml:"shutdown_timeout_sec"`
	ReadTimeoutSec     int    `yaml:"read_timeout_sec"`
	WriteTimeoutSec    int    `yaml:"write_timeout_sec"`
}

type DatabaseConfig struct {
	DSN                  string `yaml:"dsn"`
	Slaves               string `yaml:"slaves"`
	MaxOpenConns         int    `yaml:"max_open_conns"`
	MaxIdleConns         int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSec   int    `yaml:"conn_max_lifetime_sec"`
	ConnectRetries       int    `yaml:"connect_retries"`
	ConnectRetryDelaySec int    `yaml:"connect_retry_delay_sec"`
}

type MigrationsConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

func Load(path string) (*Config, error) {
	cfgw := wbfconf.New()

	if err := cfgw.LoadConfigFiles(path); err != nil {
		return nil, fmt.Errorf("load config from %q: %w", path, err)
	}

	var cfg Config
	if err := cfgw.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	zlog.Logger.Info().
		Str("dsn", cfg.Database.DSN).
		Msg("config loaded")

	if strings.TrimSpace(cfg.Database.DSN) == "" {
		return nil, errors.New("database.dsn is required (set in config file or DATABASE_DSN env)")
	}

	return &cfg, nil
}

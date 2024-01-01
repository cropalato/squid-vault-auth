//
// conf.go
// Copyright (C) 2023 rmelo <Ricardo Melo <rmelo@ludia.com>>
//
// Distributed under terms of the MIT license.
//

package conf

import (

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// ErrMissingEnvironmentAdminID missing admin_id configuration
	ErrMissingEnvironmentAdminID = errors.New("Missing Admin_id ENV Variable")

	// ErrMissingEnvironmentAdminSecret missing admin_secret configuration
	ErrMissingEnvironmentAdminSecret = errors.New("Missing Admin_secret ENV Variable")
)

// Config for the environment
type Config struct {
    Debug       bool   `envconfig:"DEBUG"`
    Addr        string `envconfig:"ADDR" default:":8080"`
    AdminID     string `envconfig:"ADMIN_ID"`
    AdminSecret string `envconfig:"ADMIN_SECRET"`
    DbPath      string `envconfig:"DB_PATH" default:"/etc/squid-vault.json"`
    CorsOrigin  string `envconfig:"CORS_ORIGIN" default:"*"`
}


func (cfg *Config) validate() error {
	if cfg.AdminID == "" {
		return ErrMissingEnvironmentAdminID
	}
	if cfg.AdminSecret == "" {
		return ErrMissingEnvironmentAdminSecret
	}

	return nil
}

func (cfg *Config) logging() error {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return nil
}

// NewDefaultConfig reads configuration from environment variables and validates it
func NewDefaultConfig() (*Config, error) {
	cfg := new(Config)
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse environment config")
	}
	err = cfg.validate()
	if err != nil {
		return nil, errors.Wrap(err, "failed validation of config")
	}
	err = cfg.logging()
	if err != nil {
		return nil, errors.Wrap(err, "failed setup logging based on config")
	}
	log.Info().Str("stage", cfg.AdminID).Bool("debug", cfg.Debug).Msg("logging configured")
	log.Info().Str("stage", cfg.AdminID).Str("branch", cfg.AdminID).Msg("Configuration loaded")

	return cfg, nil
}

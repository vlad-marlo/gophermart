package config

import (
	"flag"
	"fmt"
	"testing"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	BindAddr             string `env:"RUN_ADDRESS" envDefault:":8000"`
	DBURI                string `env:"DATABASE_URI"`
	AccuralSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
}

func New() (Config, error) {
	c := Config{}

	if err := env.Parse(&c); err != nil {
		return Config{}, fmt.Errorf("env parse: %v", err)
	}
	// parse flags
	flag.StringVar(&c.BindAddr, "a", c.BindAddr, "address to run HTTP server")
	flag.StringVar(&c.DBURI, "d", c.DBURI, "database URI")
	flag.StringVar(&c.AccuralSystemAddress, "r", c.AccuralSystemAddress, "accural system address")
	flag.Parse()

	if len(c.DBURI) == 0 {
		return Config{}, ErrEmptyDataBaseURI
	}
	if len(c.AccuralSystemAddress) == 0 {
		return Config{}, ErrEmptyDataBaseURI
	}
	return c, nil
}

func TestConfig(t *testing.T) *Config {
	c := &Config{}

	if err := env.Parse(c); err != nil {
		t.Fatalf("env parse: %v", err)
	}
	return c
}

package config

import (
	"errors"
	"os"
)

type Config struct {
	YNABAPIKey       string
	GoCardlessAPIKey string
}

func NewFromEnv() (*Config, error) {
	c := Config{
		YNABAPIKey:       os.Getenv("YAGI_YNAB_API_KEY"),
		GoCardlessAPIKey: os.Getenv("YAGI_GOCARDLESS_API_KEY"),
	}

	if c.YNABAPIKey == "" {
		return nil, errors.New("YNAB_API_KEY is not set")
	}
	if c.GoCardlessAPIKey == "" {
		return nil, errors.New("YAGI_GOCARDLESS_API_KEY is not set")
	}

	return &c, nil
}

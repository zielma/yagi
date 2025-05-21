package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	YNABAPIKey             string
	GoCardlessClientID     string
	GoCardlessClientSecret string
	YNABBaseURL            string
}

const (
	EnvYnabApiKey             = "YAGI_YNAB_API_KEY"
	EnvGoCardlessClientID     = "YAGI_GOCARDLESS_CLIENT_ID"
	EnvGoCardlessClientSecret = "YAGI_GOCARDLESS_CLIENT_SECRET"
	EnvYnabBaseURL            = "YAGI_YNAB_BASE_URL" // New environment variable for YNAB Base URL
	DefaultYnabBaseURL        = "https://api.ynab.com/v1" // Default YNAB API URL
)

var (
	ErrYNABAPIKeyNotSet             = errors.New(EnvYnabApiKey + " is not set")
	ErrGoCardlessClientIDNotSet     = errors.New(EnvGoCardlessClientID + " is not set")
	ErrGoCardlessClientSecretNotSet = errors.New(EnvGoCardlessClientSecret + " is not set")
)

func NewFromEnv() (*Config, error) {
	c := Config{
		YNABAPIKey:             os.Getenv(EnvYnabApiKey),
		GoCardlessClientID:     os.Getenv(EnvGoCardlessClientID),
		GoCardlessClientSecret: os.Getenv(EnvGoCardlessClientSecret),
		YNABBaseURL:            os.Getenv(EnvYnabBaseURL),
	}

	if strings.TrimSpace(c.YNABBaseURL) == "" {
		c.YNABBaseURL = DefaultYnabBaseURL
	}

	if strings.TrimSpace(c.YNABAPIKey) == "" {
		return nil, ErrYNABAPIKeyNotSet
	}
	if strings.TrimSpace(c.GoCardlessClientID) == "" {
		return nil, ErrGoCardlessClientIDNotSet
	}
	if strings.TrimSpace(c.GoCardlessClientSecret) == "" {
		return nil, ErrGoCardlessClientSecretNotSet
	}

	return &c, nil
}

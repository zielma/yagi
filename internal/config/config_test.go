package config_test

import (
	"os"
	"testing"

	"github.com/zielma/yagi/internal/config"
)

func TestNewFromEnv(t *testing.T) {
	os.Setenv(config.EnvYnabApiKey, "test")
	os.Setenv(config.EnvGoCardlessClientID, "test")
	os.Setenv(config.EnvGoCardlessClientSecret, "test")

	t.Run("successfully returns config", func(t *testing.T) {
		c, err := config.NewFromEnv()
		if err != nil {
			t.Fatal(err)
		}

		if c.YNABAPIKey != "test" {
			t.Fatal(config.EnvYnabApiKey + " is invalid, got " + c.YNABAPIKey)
		}

		if c.GoCardlessClientID != "test" {
			t.Fatal(config.EnvGoCardlessClientID + " is invalid, got " + c.GoCardlessClientID)
		}

		if c.GoCardlessClientSecret != "test" {
			t.Fatal(config.EnvGoCardlessClientSecret + " is invalid, got " + c.GoCardlessClientSecret)
		}
	})

	t.Run("should return error if "+config.EnvYnabApiKey+" is not set", func(t *testing.T) {
		os.Unsetenv(config.EnvYnabApiKey)
		os.Setenv(config.EnvGoCardlessClientID, "test")
		os.Setenv(config.EnvGoCardlessClientSecret, "test")
		_, err := config.NewFromEnv()
		if err != config.ErrYNABAPIKeyNotSet {
			t.Fatal("Expected ErrYNABAPIKeyNotSet, got", err)
		}
	})

	t.Run("should return error if "+config.EnvGoCardlessClientID+" is not set", func(t *testing.T) {
		os.Setenv(config.EnvYnabApiKey, "test")
		os.Unsetenv(config.EnvGoCardlessClientID)
		os.Setenv(config.EnvGoCardlessClientSecret, "test")
		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientIDNotSet {
			t.Fatal("Expected ErrGoCardlessClientIDNotSet, got", err)
		}
	})

	t.Run("should return error if "+config.EnvGoCardlessClientSecret+" is not set", func(t *testing.T) {
		os.Setenv(config.EnvYnabApiKey, "test")
		os.Setenv(config.EnvGoCardlessClientID, "test")
		os.Unsetenv(config.EnvGoCardlessClientSecret)
		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientSecretNotSet {
			t.Fatal("Expected ErrGoCardlessClientSecretNotSet, got", err)
		}
	})

	t.Run("empty "+config.EnvYnabApiKey, func(t *testing.T) {
		os.Setenv(config.EnvYnabApiKey, "")
		os.Setenv(config.EnvGoCardlessClientID, "test")
		os.Setenv(config.EnvGoCardlessClientSecret, "test")
		_, err := config.NewFromEnv()
		if err != config.ErrYNABAPIKeyNotSet {
			t.Fatal("Expected ErrYNABAPIKeyNotSet, got", err)
		}
	})

	t.Run("empty "+config.EnvGoCardlessClientID, func(t *testing.T) {
		os.Setenv(config.EnvYnabApiKey, "test")
		os.Setenv(config.EnvGoCardlessClientID, "")
		os.Setenv(config.EnvGoCardlessClientSecret, "test")
		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientIDNotSet {
			t.Fatal("Expected ErrGoCardlessClientIDNotSet, got", err)
		}
	})

	t.Run("empty "+config.EnvGoCardlessClientSecret, func(t *testing.T) {
		os.Setenv(config.EnvYnabApiKey, "test")
		os.Setenv(config.EnvGoCardlessClientID, "test")
		os.Setenv(config.EnvGoCardlessClientSecret, "")
		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientSecretNotSet {
			t.Fatal("Expected ErrGoCardlessClientSecretNotSet, got", err)
		}
	})
}

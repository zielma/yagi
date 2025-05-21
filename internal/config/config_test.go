package config_test

import (
	"os"
	"testing"

	"github.com/zielma/yagi/internal/config"
)

func TestNewFromEnv(t *testing.T) {
	// Helper function to set common environment variables for tests
	setCommonEnv := func() {
		os.Setenv(config.EnvYnabApiKey, "test-key")
		os.Setenv(config.EnvGoCardlessClientID, "test-client-id")
		os.Setenv(config.EnvGoCardlessClientSecret, "test-client-secret")
	}

	// Helper function to unset common environment variables after tests
	unsetCommonEnv := func() {
		os.Unsetenv(config.EnvYnabApiKey)
		os.Unsetenv(config.EnvGoCardlessClientID)
		os.Unsetenv(config.EnvGoCardlessClientSecret)
		os.Unsetenv(config.EnvYnabBaseURL) // Ensure YNABBaseURL is also cleaned up
	}

	t.Run("successfully returns config with default YNABBaseURL", func(t *testing.T) {
		setCommonEnv()
		os.Unsetenv(config.EnvYnabBaseURL) // Ensure YNABBaseURL is not set to get default
		defer unsetCommonEnv()

		c, err := config.NewFromEnv()
		if err != nil {
			t.Fatal(err)
		}

		if c.YNABAPIKey != "test-key" {
			t.Errorf("Expected YNABAPIKey 'test-key', got %s", c.YNABAPIKey)
		}
		if c.GoCardlessClientID != "test-client-id" {
			t.Errorf("Expected GoCardlessClientID 'test-client-id', got %s", c.GoCardlessClientID)
		}
		if c.GoCardlessClientSecret != "test-client-secret" {
			t.Errorf("Expected GoCardlessClientSecret 'test-client-secret', got %s", c.GoCardlessClientSecret)
		}
		if c.YNABBaseURL != config.DefaultYnabBaseURL {
			t.Errorf("Expected YNABBaseURL '%s', got %s", config.DefaultYnabBaseURL, c.YNABBaseURL)
		}
	})

	t.Run("successfully returns config with custom YNABBaseURL", func(t *testing.T) {
		setCommonEnv()
		customURL := "http://localhost:9090/api/v1"
		os.Setenv(config.EnvYnabBaseURL, customURL)
		defer unsetCommonEnv()

		c, err := config.NewFromEnv()
		if err != nil {
			t.Fatal(err)
		}

		if c.YNABAPIKey != "test-key" {
			t.Errorf("Expected YNABAPIKey 'test-key', got %s", c.YNABAPIKey)
		}
		if c.YNABBaseURL != customURL {
			t.Errorf("Expected YNABBaseURL '%s', got %s", customURL, c.YNABBaseURL)
		}
	})

	t.Run("successfully returns config with empty YNABBaseURL triggering default", func(t *testing.T) {
		setCommonEnv()
		os.Setenv(config.EnvYnabBaseURL, "  ") // Set to empty string with spaces
		defer unsetCommonEnv()

		c, err := config.NewFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		if c.YNABBaseURL != config.DefaultYnabBaseURL {
			t.Errorf("Expected YNABBaseURL '%s' for empty env var, got %s", config.DefaultYnabBaseURL, c.YNABBaseURL)
		}
	})

	t.Run("should return error if "+config.EnvYnabApiKey+" is not set", func(t *testing.T) {
		// setCommonEnv() // Don't set YNABApiKey
		os.Setenv(config.EnvGoCardlessClientID, "test-client-id")
		os.Setenv(config.EnvGoCardlessClientSecret, "test-client-secret")
		os.Unsetenv(config.EnvYnabApiKey)
		defer unsetCommonEnv()

		_, err := config.NewFromEnv()
		if err != config.ErrYNABAPIKeyNotSet {
			t.Fatalf("Expected ErrYNABAPIKeyNotSet, got %v", err)
		}
	})

	t.Run("should return error if "+config.EnvGoCardlessClientID+" is not set", func(t *testing.T) {
		// setCommonEnv() // Don't set GoCardlessClientID
		os.Setenv(config.EnvYnabApiKey, "test-key")
		os.Setenv(config.EnvGoCardlessClientSecret, "test-client-secret")
		os.Unsetenv(config.EnvGoCardlessClientID)
		defer unsetCommonEnv()

		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientIDNotSet {
			t.Fatalf("Expected ErrGoCardlessClientIDNotSet, got %v", err)
		}
	})

	t.Run("should return error if "+config.EnvGoCardlessClientSecret+" is not set", func(t *testing.T) {
		// setCommonEnv() // Don't set GoCardlessClientSecret
		os.Setenv(config.EnvYnabApiKey, "test-key")
		os.Setenv(config.EnvGoCardlessClientID, "test-client-id")
		os.Unsetenv(config.EnvGoCardlessClientSecret)
		defer unsetCommonEnv()
		
		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientSecretNotSet {
			t.Fatalf("Expected ErrGoCardlessClientSecretNotSet, got %v", err)
		}
	})

	// Tests for empty values that should trigger errors
	t.Run("empty "+config.EnvYnabApiKey, func(t *testing.T) {
		setCommonEnv() // Set all, then override one to be empty
		os.Setenv(config.EnvYnabApiKey, "")
		defer unsetCommonEnv()

		_, err := config.NewFromEnv()
		if err != config.ErrYNABAPIKeyNotSet {
			t.Fatalf("Expected ErrYNABAPIKeyNotSet for empty key, got %v", err)
		}
	})

	t.Run("empty "+config.EnvGoCardlessClientID, func(t *testing.T) {
		setCommonEnv()
		os.Setenv(config.EnvGoCardlessClientID, "")
		defer unsetCommonEnv()

		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientIDNotSet {
			t.Fatalf("Expected ErrGoCardlessClientIDNotSet for empty ID, got %v", err)
		}
	})

	t.Run("empty "+config.EnvGoCardlessClientSecret, func(t *testing.T) {
		setCommonEnv()
		os.Setenv(config.EnvGoCardlessClientSecret, "")
		defer unsetCommonEnv()

		_, err := config.NewFromEnv()
		if err != config.ErrGoCardlessClientSecretNotSet {
			t.Fatalf("Expected ErrGoCardlessClientSecretNotSet for empty secret, got %v", err)
		}
	})
}

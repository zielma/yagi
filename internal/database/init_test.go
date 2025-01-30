package database

import (
	"fmt"
	"os"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Initial setup for testing
	testDataFolder := "data_test"
	testMigrationsFolder := "migrations_test"

	if err := os.Mkdir(testDataFolder, 0755); err != nil {
		t.Fatalf("Failed to create data folder: %v", err)
	}
	if err := os.Mkdir(testMigrationsFolder, 0755); err != nil {
		t.Fatalf("Failed to create migrations folder: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(testDataFolder); err != nil {
			t.Fatalf("Failed to remove data folder: %v", err)
		}

		if err := os.RemoveAll(testMigrationsFolder); err != nil {
			t.Fatalf("Failed to remove migrations folder: %v", err)
		}
	})

	t.Run("invalid migrations folder", func(t *testing.T) {
		// Act
		_, err := initialize(dataFolder, dbFileName, "invalid_folder")

		// Assert
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("successful initialization", func(t *testing.T) {
		// Act
		db, err := Initialize()

		//Assert
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if db == nil {
			t.Fatal("Expected database connection, got nil")
		}
		defer db.Close()

		var result int
		err = db.QueryRow("SELECT 1").Scan(&result)
		if err != nil {
			t.Fatalf("Failed to query database: %v", err)
		}
		if result != 1 {
			t.Errorf("Expected 1, got %d", result)
		}
	})

	t.Run("failed migrations", func(t *testing.T) {
		// Arrange
		migration := []byte("drop table test;")
		if err := os.WriteFile(fmt.Sprintf("%s/%s", testMigrationsFolder, "000001_initial.up.sql"), migration, 0644); err != nil {
			t.Fatalf("Failed to write migration file: %v", err)
		}

		// Act
		db, err := initialize(testDataFolder, dbFileName, testMigrationsFolder)

		// Assert
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if db != nil {
			t.Fatal("Expected nil database connection, got non-nil")
		}
	})
}

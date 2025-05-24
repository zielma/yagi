package database

import (
	"fmt"
	"os"
	"testing"
)

const (
	testDataFolder       = "data_test"
	testDbFileName       = "yagi_test.sqlite3"
	testMigrationsFolder = "migrations_test"
)

func setupTest(t *testing.T) {
	t.Helper()
	if err := os.Mkdir(testDataFolder, 0755); err != nil {
		t.Fatalf("Failed to create data folder: %v", err)
	}
	if err := os.Mkdir(testMigrationsFolder, 0755); err != nil {
		t.Fatalf("Failed to create migrations folder: %v", err)
	}
}

func cleanup(t *testing.T) {
	if err := os.RemoveAll(testDataFolder); err != nil {
		t.Fatalf("Failed to remove data folder: %v", err)
	}

	if err := os.RemoveAll(testMigrationsFolder); err != nil {
		t.Fatalf("Failed to remove migrations folder: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	t.Run("data folder doesn't exists", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { cleanup(t) })

		// Act
		_, err := initialize(testDataFolder, dbFileName, migrationsFolder)

		// Assert
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
	})

	t.Run("migrations folder doesn't exists", func(t *testing.T) {
		// Arrange
		setupTest(t)
		t.Cleanup(func() { cleanup(t) })

		// Act
		_, err := initialize(testDataFolder, dbFileName, "invalid_folder")

		// Assert
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("invalid migration script", func(t *testing.T) {
		// Arrange
		setupTest(t)
		t.Cleanup(func() { cleanup(t) })

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

	t.Run("successful initialization", func(t *testing.T) {
		// Arrange
		setupTest(t)
		t.Cleanup(func() { cleanup(t) })

		// Act
		db, err := initialize(testDataFolder, dbFileName, migrationsFolder)

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
}

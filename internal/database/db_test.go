package database

import (
	"context"
	"database/sql"
	"testing"
)

type mockDBTX struct{}

func (m *mockDBTX) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}
func (m *mockDBTX) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockDBTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

func TestNew(t *testing.T) {
	// Setup mock
	mockDB := &mockDBTX{}

	// Get queries instance
	queries := New(mockDB)

	// Validate queries use the mockDB
	if queries == nil {
		t.Fatal("Expected New() to return non-nil Queries instance")
	}
	if queries.db != mockDB {
		t.Error("Expected db field to be set to mockDB")
	}
}

func TestWithTx(t *testing.T) {
	// Setup mocks
	mockDB := &mockDBTX{}
	queries := New(mockDB)
	tx := &sql.Tx{}

	// Get queries with transaction
	txQueries := queries.WithTx(tx)

	// Validate that queries use the transaction
	if txQueries == nil {
		t.Fatal("Expected WithTx() to return non-nil Queries instance")
	}
	if txQueries.db != tx {
		t.Error("Expected db field to be set to the transaction")
	}
}

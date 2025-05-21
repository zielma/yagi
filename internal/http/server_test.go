package http_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	internalHTTP "github.com/zielma/yagi/internal/http"
	"github.com/zielma/yagi/internal/router"
)

func TestErrorResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	// Use the now exported NewErrorResponseWriter
	erw := internalHTTP.NewErrorResponseWriter(rr)

	erw.WriteHeader(http.StatusNotFound)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected underlying ResponseWriter status code %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestErrorResponseWriter_Write_StatusOk(t *testing.T) {
	rr := httptest.NewRecorder()
	erw := internalHTTP.NewErrorResponseWriter(rr)
	testBody := []byte("test body")

	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	defer slog.SetDefault(originalLogger)

	erw.WriteHeader(http.StatusOK)
	_, err := erw.Write(testBody)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if rr.Body.String() != string(testBody) {
		t.Errorf("expected body %q, got %q", string(testBody), rr.Body.String())
	}

	logContent := logOutput.String()
	if strings.Contains(logContent, "error response") {
		t.Errorf("expected log output not to contain 'error response', but got %q", logContent)
	}
}

func TestErrorResponseWriter_Write_StatusError(t *testing.T) {
	rr := httptest.NewRecorder()
	erw := internalHTTP.NewErrorResponseWriter(rr)
	testBody := []byte("error message")

	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	defer slog.SetDefault(originalLogger)

	erw.WriteHeader(http.StatusInternalServerError)
	_, err := erw.Write(testBody)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if rr.Body.String() != string(testBody) {
		t.Errorf("expected body %q, got %q", string(testBody), rr.Body.String())
	}

	logContent := logOutput.String()
	if !strings.Contains(logContent, "error response") {
		t.Errorf("expected log output to contain 'error response', got %q", logContent)
	}
	if !strings.Contains(logContent, "status_code=500") {
		t.Errorf("expected log output to contain 'status_code=500', got %q", logContent)
	}
	if !strings.Contains(logContent, "error=\"error message\"") {
		t.Errorf("expected log output to contain 'error=\"error message\"', got %q", logContent)
	}
}

func TestNewServer_LoggingMiddlewareUsesErrorResponseWriter(t *testing.T) {
	r := router.NewRouter()
	server := internalHTTP.NewServer(r)

	r.ServeMux.HandleFunc("/testerror", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("handler causing error"))
	})

	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	defer slog.SetDefault(originalLogger)

	req := httptest.NewRequest(http.MethodGet, "/testerror", nil)
	rr := httptest.NewRecorder()

	server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if rr.Body.String() != "handler causing error" {
		t.Errorf("expected body %q, got %q", "handler causing error", rr.Body.String())
	}

	logContent := logOutput.String()
	if !strings.Contains(logContent, "error response") {
		t.Errorf("expected server log output to contain 'error response', got %q", logContent)
	}
	if !strings.Contains(logContent, "status_code=400") {
		t.Errorf("expected server log output to contain 'status_code=400', got %q", logContent)
	}
	if !strings.Contains(logContent, "error=\"handler causing error\"") {
		t.Errorf("expected server log output to contain 'error=\"handler causing error\"', got %q", logContent)
	}
}

func TestNewServer_HandlerWorksForNonError(t *testing.T) {
	r := router.NewRouter()
	server := internalHTTP.NewServer(r)

	r.ServeMux.HandleFunc("/testok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	var logOutput bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logOutput, nil)))
	defer slog.SetDefault(originalLogger)

	req := httptest.NewRequest(http.MethodGet, "/testok", nil)
	rr := httptest.NewRecorder()

	server.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got %q", rr.Body.String())
	}

	logContent := logOutput.String()
	if strings.Contains(logContent, "error response") {
		t.Errorf("expected log output not to contain 'error response' for OK request, but got %q", logContent)
	}
}

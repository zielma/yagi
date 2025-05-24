package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zielma/yagi/internal/router"
)

type mockResponseWriter struct {
	headers    http.Header
	body       bytes.Buffer
	statusCode int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		headers: make(http.Header),
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestNewErrorLoggingResponseWriter(t *testing.T) {
	baseWriter := newMockResponseWriter()
	ew := newErrorLoggingResponseWriter(baseWriter)

	if ew.statusCode != http.StatusOK {
		t.Errorf("Expected initial status code %d, got %d", http.StatusOK, ew.statusCode)
	}
	if ew.ResponseWriter != baseWriter {
		t.Error("ResponseWriter not properly set")
	}
	if ew.buf == nil {
		t.Error("Buffer should not be nil")
	}
}

func TestErrorLoggingResponseWriter_WriteHeader(t *testing.T) {
	baseWriter := newMockResponseWriter()
	ew := newErrorLoggingResponseWriter(baseWriter)

	testCode := http.StatusBadRequest
	ew.WriteHeader(testCode)

	if ew.statusCode != testCode {
		t.Errorf("Expected status code %d, got %d", testCode, ew.statusCode)
	}
	if baseWriter.statusCode != testCode {
		t.Errorf("Base writer status code not set, expected %d, got %d", testCode, baseWriter.statusCode)
	}
}

func TestErrorLoggingResponseWriter_Write(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       []byte
		shouldLog  bool
	}{
		{
			name:       "success response",
			statusCode: http.StatusOK,
			data:       []byte("success"),
			shouldLog:  false,
		},
		{
			name:       "error response",
			statusCode: http.StatusBadRequest,
			data:       []byte("error"),
			shouldLog:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseWriter := newMockResponseWriter()
			ew := newErrorLoggingResponseWriter(baseWriter)
			ew.WriteHeader(tt.statusCode)

			n, err := ew.Write(tt.data)

			if err != nil {
				t.Errorf("Write returned unexpected error: %v", err)
			}
			if n != len(tt.data) {
				t.Errorf("Write returned wrong length: got %d, want %d", n, len(tt.data))
			}
			if tt.shouldLog && ew.buf.String() != string(tt.data) {
				t.Errorf("Error not logged correctly, got %q, want %q", ew.buf.String(), string(tt.data))
			}
			if !tt.shouldLog && ew.buf.Len() > 0 {
				t.Error("Buffer should be empty for success responses")
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	// Setup router with a test handler
	r := router.New()
	errorMsg := "test error"
	r.ServeMux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errorMsg, http.StatusBadRequest)
	})
	r.ServeMux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	server := NewServer(r)

	if server.Addr != ":8080" {
		t.Errorf("Expected server address :8080, got %s", server.Addr)
	}
	if server.Handler == nil {
		t.Error("Server handler should not be nil")
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
		shouldLog      bool
	}{
		{
			name:           "success response",
			path:           "/success",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
			shouldLog:      false,
		},
		{
			name:           "error response",
			path:           "/error",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errorMsg + "\n", // http.Error adds a newline
			shouldLog:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			server.Handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if body := w.Body.String(); body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

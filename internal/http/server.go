package http

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/zielma/yagi/internal/router"
)

type errorResponseWriter struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func newErrorResponseWriter(w http.ResponseWriter) *errorResponseWriter {
	return &errorResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		buf:            &bytes.Buffer{},
	}
}

func (w *errorResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode >= 400 {
		w.buf.Write(b)
		slog.Warn("error response",
			"status_code", w.statusCode,
			"error", string(b),
		)
	}
	return w.ResponseWriter.Write(b)
}

func NewServer(router *router.Router) *http.Server {
	logWrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew := newErrorResponseWriter(w)
		router.ServeMux.ServeHTTP(ew, r)
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: logWrapper,
	}
}

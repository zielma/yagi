package router

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestBasicRouting(t *testing.T) {
	r := NewRouter()
	called := false

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Error("handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("status code, want: %d, got: %d", http.StatusOK, w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := NewRouter()
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code, want: %d, got: %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestMiddleware(t *testing.T) {
	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	r := NewRouter(middleware)
	handlerCalled := false

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if _, err := w.Write([]byte("hello")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("middleware was not called")
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}

	if w.Body.String() != "hello" {
		t.Errorf("response body, want: %s, got: %s", "hello", w.Body.String())
	}
}

func TestGroup(t *testing.T) {
	r := NewRouter()
	groupCalled := false
	middlewareCalled := false

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	r.Group(func(g *Router) {
		g.Use(middleware)
		g.Get("/group", func(w http.ResponseWriter, r *http.Request) {
			groupCalled = true
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/group", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("group middleware was not called")
	}
	if !groupCalled {
		t.Error("group handler was not called")
	}
}

func TestMultipleMiddleware(t *testing.T) {
	order := []string{}

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2")
			next.ServeHTTP(w, r)
		})
	}

	r := NewRouter(middleware1)
	r.Use(middleware2)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expected := []string{"m1", "m2", "handler"}
	if !slices.Equal(order, expected) {
		t.Errorf("incorrect middleware execution order, want: %#v, got: %#v", expected, order)
	}
}

func TestAllHttpMethods(t *testing.T) {

	r := NewRouter()
	tests := []struct {
		method string
		setup  func(string, http.HandlerFunc, ...Middleware)
	}{
		{http.MethodGet, r.Get},
		{http.MethodPost, r.Post},
		{http.MethodPut, r.Put},
		{http.MethodDelete, r.Delete},
		{http.MethodHead, r.Head},
		{http.MethodOptions, r.Options},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			called := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
			}

			tt.setup("/test", handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if !called {
				t.Errorf("handler[%s] was not called", tt.method)
			}
		})
	}
}

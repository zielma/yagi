package router

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestRouter_BasicRouting(t *testing.T) {
	r := NewRouter()
	called := false

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := NewRouter()
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestRouter_Middleware(t *testing.T) {
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
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouter_Group(t *testing.T) {
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
		t.Error("Group middleware was not called")
	}
	if !groupCalled {
		t.Error("Group handler was not called")
	}
}

func TestRouter_MultipleMiddleware(t *testing.T) {
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
		t.Errorf("Expected middleware execution order %v, got %v", expected, order)
	}
}

func TestRouter_AllMethods(t *testing.T) {

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
				t.Errorf("%s handler was not called", tt.method)
			}
		})
	}
}

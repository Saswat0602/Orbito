package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`"Welcome to Orbito Backend API!"`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	expected := `"Welcome to Orbito Backend API!"`
	if rr.Body.String() != expected {
		t.Fatalf("expected %q, got %q", expected, rr.Body.String())
	}
}

package main

import (
	"issue-pm/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAndFrontendRoutes(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test", JWTExpireHours: 1}
	r := newRouter(cfg, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() == "" {
		t.Fatalf("health failed: %d %s", w.Code, w.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("frontend route failed: %d", w.Code)
	}
}

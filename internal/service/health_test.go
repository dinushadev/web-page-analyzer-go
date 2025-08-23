package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	HealthCheckHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

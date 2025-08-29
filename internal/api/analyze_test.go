package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"test-project-go/internal/model"
	"test-project-go/internal/service"
)

func TestAnalyzeHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	rr := httptest.NewRecorder()
	AnalyzeHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAnalyzeHandler_BadRequestBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	AnalyzeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAnalyzeHandler_InvalidURL(t *testing.T) {
	old := analyzePageFunc
	defer func() { analyzePageFunc = old }()
	analyzePageFunc = func(ctx context.Context, u string) (*model.AnalyzeResult, error) {
		return nil, service.ErrInvalidURL
	}
	body, _ := json.Marshal(map[string]string{"url": ":bad:"})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	AnalyzeHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAnalyzeHandler_UpstreamError(t *testing.T) {
	old := analyzePageFunc
	defer func() { analyzePageFunc = old }()
	analyzePageFunc = func(ctx context.Context, u string) (*model.AnalyzeResult, error) {
		return nil, service.ErrUpstream
	}
	body, _ := json.Marshal(map[string]string{"url": "http://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	AnalyzeHandler(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rr.Code)
	}
}

func TestAnalyzeHandler_Success(t *testing.T) {
	old := analyzePageFunc
	defer func() { analyzePageFunc = old }()
	analyzePageFunc = func(ctx context.Context, u string) (*model.AnalyzeResult, error) {
		return &model.AnalyzeResult{Title: "ok"}, nil
	}
	body, _ := json.Marshal(map[string]string{"url": "http://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	AnalyzeHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var res model.AnalyzeResult
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if res.Title != "ok" {
		t.Fatalf("unexpected title: %q", res.Title)
	}
}

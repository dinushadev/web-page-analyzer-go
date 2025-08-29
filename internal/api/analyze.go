package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"test-project-go/internal/analyzer"
)

var analyzePageFunc = analyzer.AnalyzePage

type analyzeRequest struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

// AnalyzeHandler handles the analysis of a web page.
// @Summary Analyze a web page
// @Description Analyzes the given URL and returns HTML version, title, headings, link stats, and login form presence.
// @Tags analyze
// @Accept json
// @Produce json
// @Param analyzeRequest body analyzeRequest true "URL to analyze"
// @Success 200 {object} model.AnalyzeResult
// @Failure 400 {object} errorResponse
// @Failure 502 {object} errorResponse
// @Router /analyze [post]
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Invalid request body", StatusCode: http.StatusBadRequest})
		return
	}
	result, err := analyzePageFunc(r.Context(), req.URL)
	if err != nil {
		code := http.StatusBadGateway
		switch {
		case errors.Is(err, analyzer.ErrInvalidURL):
			code = http.StatusBadRequest
		case errors.Is(err, analyzer.ErrUnreachable), errors.Is(err, analyzer.ErrUpstream):
			code = http.StatusBadGateway
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error(), StatusCode: code})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

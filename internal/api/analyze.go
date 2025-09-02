package api

import (
	"encoding/json"
	"net/http"
	"web-analyzer-go/internal/analyzer"
	appErr "web-analyzer-go/internal/errors"
)

var analyzePageFunc = analyzer.AnalyzePage

type analyzeRequest struct {
	URL string `json:"url"`
}

// errorResponse is deprecated, use appErr.HTTPResponse instead
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
// @Failure 400 {object} errors.HTTPResponse
// @Failure 502 {object} errors.HTTPResponse
// @Router /analyze [post]
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		appErr.HTTPErrorHandler(w, r, appErr.NewMethodNotAllowedError([]string{http.MethodPost}))
		return
	}

	// Limit request body size to 1MB and ensure close
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr.HTTPErrorHandler(w, r, appErr.NewBadRequestError("Invalid request body", err))
		return
	}
	if req.URL == "" {
		appErr.HTTPErrorHandler(w, r, appErr.NewValidationError("URL is required"))
		return
	}
	result, err := analyzePageFunc(r.Context(), req.URL)
	if err != nil {
		appErr.HTTPErrorHandler(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

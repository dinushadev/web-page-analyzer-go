package api

import (
	"encoding/json"
	"net/http"
	"test-project-go/internal/service"
)

type analyzeRequest struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "Invalid request body", StatusCode: 400})
		return
	}
	result, status, err := service.AnalyzePage(req.URL)
	if err != nil {
		code := status
		if code == 0 {
			code = http.StatusBadGateway
		}
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error(), StatusCode: code})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

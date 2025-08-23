package api

import (
	"net/http"
	"test-project-go/internal/service"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.HealthCheckHandler)
	return mux
}

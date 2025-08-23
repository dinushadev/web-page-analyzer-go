package api

import (
	"net/http"
	"test-project-go/internal/service"
	"test-project-go/internal/middleware"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.HealthCheckHandler)
	return middleware.Logging(mux)
}

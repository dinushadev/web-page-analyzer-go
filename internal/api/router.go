package api

import (
	"net/http"
	"test-project-go/internal/service"
	"test-project-go/internal/middleware"
	"test-project-go/internal/metrics"
	_ "net/http/pprof"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.HealthCheckHandler)
	mux.HandleFunc("/analyze", AnalyzeHandler)
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	return middleware.Logging(mux)
}

package api

import (
	"net/http"
	"test-project-go/internal/service"
	"test-project-go/internal/middleware"
	"test-project-go/internal/metrics"
	_ "net/http/pprof"
	"github.com/swaggo/http-swagger"
	_ "test-project-go/docs"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.HealthCheckHandler)
	mux.HandleFunc("/analyze", AnalyzeHandler)
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	// Serve static files from the web directory at the root path
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return middleware.Logging(mux)
}

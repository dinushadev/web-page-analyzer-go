package api

import (
	"net/http"
	_ "net/http/pprof"
	_ "test-project-go/docs"
	"test-project-go/internal/metrics"
	"test-project-go/internal/middleware"
	"test-project-go/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
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
	// Chain: RequestID -> Recover -> Logging -> mux
	return middleware.RequestID(middleware.Recover(middleware.Logging(mux)))
}

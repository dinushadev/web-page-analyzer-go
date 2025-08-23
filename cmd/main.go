package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test-project-go/internal/api"
	"test-project-go/internal/metrics"
	"test-project-go/internal/util"

	_ "net/http/pprof"
)

func main() {
	util.InitLogger()
	metrics.RegisterPrometheus()

	mux := api.NewRouter()

	// Metrics and pprof endpoints
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Server started on :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}

package main

import (
	"context"
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

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()
	util.Logger.Info("server.started", "addr", ":8081")

	<-ctx.Done()
	util.Logger.Info("server.shutting_down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		util.Logger.Error("server.shutdown_forced", "error", err)
		os.Exit(1)
	}
	util.Logger.Info("server.exiting")
}

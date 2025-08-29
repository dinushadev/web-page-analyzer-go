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
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Logger.Error("listen", "error", err)
			os.Exit(1)
		}
	}()
	util.Logger.Info("server.started", "addr", ":8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	util.Logger.Info("server.shutting_down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		util.Logger.Error("server.shutdown_forced", "error", err)
		os.Exit(1)
	}
	util.Logger.Info("server.exiting")
}

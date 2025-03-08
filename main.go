package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VladMinzatu/go-mon/monitor"
	"github.com/VladMinzatu/go-mon/web"
)

func main() {
	systemMonitor := monitor.NewSystemMonitorService(
		monitor.NewSystemMonitor(&monitor.DefaultSystemMetricsProvider{}, 1*time.Second))
	systemMonitor.Start()

	srv, err := web.NewServer(systemMonitor)
	if err != nil {
		slog.Error("Failed to initialise server up. Shutting down", "error", err.Error())
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:    "localhost:8080",
		Handler: srv,
	}

	// Create channel for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("Starting server listening on port 8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			stop <- syscall.SIGTERM // Trigger shutdown on server start failure
		}
	}()

	// Wait for shutdown signal
	<-stop
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}

	systemMonitor.Stop()
	slog.Info("Server stopped gracefully. Exiting")
	os.Exit(0)
}

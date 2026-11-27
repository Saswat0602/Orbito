package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/logger"
)

func loadDotEnv() {
	paths := []string{".env", "../../backend/.env", "../backend/.env", "backend/.env"}
	for _, p := range paths {
		cleaned := filepath.Clean(p)
		data, err := os.ReadFile(cleaned)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		break
	}
}

func main() {
	loadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger := logger.New(cfg.IsProduction())

	var db *database.DB
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err = database.New(ctx, cfg.DatabaseURL)
		cancel()
		if err != nil {
			appLogger.Warn("Database initial connection failed, proceeding with degraded health state", slog.String("error", err.Error()))
		} else {
			defer db.Close()
		}
	}

	router, eventsGW := buildRouterWithGateway(cfg, db, appLogger)
	defer eventsGW.Close()

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		appLogger.Info("Graceful shutdown initiating...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Server forced shutdown", slog.String("error", err.Error()))
		}
		serverStopCtx()
	}()

	appLogger.Info("Orbito Go API listening", slog.String("addr", addr), slog.String("env", cfg.AppEnv))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		appLogger.Error("Server listen failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	<-serverCtx.Done()
	appLogger.Info("Server stopped cleanly")
}

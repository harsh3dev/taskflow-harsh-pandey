package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/config"
	"github.com/harshpn/taskflow/internal/httpapi"
	"github.com/harshpn/taskflow/internal/service"
	"github.com/harshpn/taskflow/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel(),
	}))

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("ping database", "error", err)
		os.Exit(1)
	}

	st := store.New(db)
	tokenManager := auth.NewTokenManager(auth.TokenManagerConfig{
		ActiveKeyID:    cfg.JWTActiveKeyID,
		SigningKeys:    cfg.JWTSigningKeys,
		AccessTokenTTL: cfg.AccessTokenTTL,
		Issuer:         cfg.JWTIssuer,
		Audience:       cfg.JWTAudience,
	})
	server := httpapi.NewServer(httpapi.Dependencies{
		Logger:              logger,
		TokenParser:         tokenManager,
		AuthService:         service.NewAuthService(st, st, tokenManager, cfg.RefreshTokenTTL, cfg.BcryptCost, nil),
		ProjectService:      service.NewProjectService(st),
		TaskService:         service.NewTaskService(st),
		UserService:         service.NewUserService(st),
		MaxRequestBodyBytes: cfg.MaxRequestBodyBytes,
	})

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", httpServer.Addr)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case serveErr := <-errCh:
		if serveErr != nil {
			logger.Error("server failed", "error", serveErr)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

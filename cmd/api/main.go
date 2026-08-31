package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devclub.com/identity/internal/api/config"
	"devclub.com/identity/internal/api/router"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/internal/mailer"
	"devclub.com/identity/pkg/utils"
)

func main() {
	startTime := time.Now()

	infoLogger, warnLogger, errorLogger := utils.CreateLoggers()

	cfg, err := config.NewConfig()
	if err != nil {
		errorLogger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pool, err := database.NewPool(dbCtx, cfg.DBConn, errorLogger)
	if err != nil {
		warnLogger.Warn("database connection or ping failed on startup, service running in degraded mode", "error", err)
	} else {
		infoLogger.Info("database pool initialized successfully")
	}
	if pool != nil {
		defer pool.Close()
	}

	redisClient, err := database.NewRedisClient(dbCtx, cfg.RedisURL, cfg.RedisURI, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB, errorLogger)
	if err != nil {
		warnLogger.Warn("redis connection or ping failed on startup, service running in degraded mode without cache", "error", err)
	} else {
		infoLogger.Info("redis client initialized successfully")
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// 4. Initialize Mailer
	mailClient, err := mailer.NewMailer(cfg, infoLogger)
	if err != nil {
		errorLogger.Error("failed to initialize mail client", "error", err)
		os.Exit(1)
	}

	// 5. Initialize JWT Manager
	jwtManager := utils.NewJWTManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
		cfg.CookieDomain,
		cfg.IsProduction,
	)

	// 6. Build Router & Wire Dependencies
	r := router.NewRouter(
		cfg,
		pool,
		redisClient,
		mailClient,
		jwtManager,
		infoLogger,
		warnLogger,
		errorLogger,
		startTime,
	)

	// 7. Setup HTTP Server
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		infoLogger.Info("server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		errorLogger.Error("server encountered a fatal error", "error", err)
		os.Exit(1)
	case sig := <-shutdownSignal:
		infoLogger.Info("shutdown signal received", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			errorLogger.Error("failed to shutdown server gracefully", "error", err)
			_ = server.Close()
		}
		infoLogger.Info("server exited cleanly")
	}
}

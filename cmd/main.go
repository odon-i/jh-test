package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forum/internal/cache"
	"forum/internal/config"
	"forum/internal/database"
	"forum/internal/router"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		exitf("load configuration: %v", err)
	}

	db, err := database.Open(cfg.Database, cfg.Server.Mode == "debug")
	if err != nil {
		exitf("database startup failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		exitf("database migration failed: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	appCache, err := cache.Open(context.Background(), cfg.Redis)
	if err != nil {
		exitf("redis startup failed: %v", err)
	}
	defer appCache.Close()

	engine := router.New(router.Dependencies{Config: cfg, DB: db, Cache: appCache})
	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownComplete := make(chan struct{})
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		close(shutdownComplete)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		exitf("server stopped unexpectedly: %v", err)
	}
	<-shutdownComplete
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

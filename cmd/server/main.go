package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/database"
	"perimatrik-api-go/internal/handlers"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/router"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/validation"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	models.Location = cfg.Location
	models.AppURL = cfg.AppURL

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	h := &handlers.Handler{
		DB:      db,
		Cfg:     cfg,
		Auth:    &auth.Service{DB: db, TokenPrefix: cfg.TokenPrefix, BcryptCost: cfg.BcryptCost},
		Storage: &storage.Storage{Root: cfg.StoragePath, AppURL: cfg.AppURL},
		V:       validation.New(db),
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(cfg, h),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("perimatrik-api listening on :%s (APP_URL=%s)", cfg.Port, cfg.AppURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

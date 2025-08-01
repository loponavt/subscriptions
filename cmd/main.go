package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"subscriptions/internal/config"
	"subscriptions/internal/handler"
	"subscriptions/internal/logger"
	"subscriptions/internal/repository"
	"subscriptions/internal/usecase"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger_ := logger.New()
	logger_.Info("Starting subscription service...")

	db, err := repository.InitDB(cfg)
	if err != nil {
		logger_.Fatalf("failed to initialize database: %v", err)
	}
	logger_.Info("Database connected and migrated")

	repo := repository.NewRepository(db)
	usc := usecase.New(repo)
	h := handler.New(usc)

	r := gin.Default()
	h.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		logger_.Infof("Listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger_.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger_.Info("Shutdown signal received, exiting...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger_.Fatalf("Server forced to shutdown: %v", err)
	}

	logger_.Info("Server exiting gracefully")
}

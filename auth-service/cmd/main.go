package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/handlers"
	"auth-service/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	_ = godotenv.Load()

	port := getenv("AUTH_SERVICE_PORT", "8081")
	secretKey := getenv("AUTH_JWT_SECRET", "default_secret")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	authSrv := service.NewAuthService(secretKey, logger)
	authHnd := handlers.NewAuthHandlers(authSrv, logger)

	r := mux.NewRouter()
	r.HandleFunc("/auth/register", authHnd.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", authHnd.LoginHandler).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("auth-service started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("server stopped")
}

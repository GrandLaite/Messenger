package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"storage-service/internal/handlers"
	"storage-service/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()

	port := getEnv("STORAGE_SERVICE_PORT", "8084")
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	minioSSL := getEnv("MINIO_SSL", "false") == "true"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	storageSvc, err := service.NewStorageService(minioEndpoint, minioAccessKey, minioSecretKey, minioSSL, logger)
	if err != nil {
		logger.Error("Failed to initialize storage service", "error", err)
		os.Exit(1)
	}

	err = storageSvc.CreateBucket(context.Background(), "docs")
	if err != nil {
		logger.Error("Failed to create bucket", "error", err)
		os.Exit(1)
	}

	storageHnd := handlers.NewStorageHandlers(storageSvc, logger)

	r := mux.NewRouter()
	r.HandleFunc("/storage/upload", storageHnd.UploadHandler).Methods(http.MethodPost)
	r.HandleFunc("/storage/download/{filename}", storageHnd.DownloadHandler).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("storage-service started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down storage-service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("storage-service stopped")
}

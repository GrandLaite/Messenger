package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cache-service/internal/handlers"
	"cache-service/internal/service"

	"github.com/gorilla/mux"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func atoi(s string) int { i, _ := strconv.Atoi(s); return i }

func main() {
	port := getenv("CACHE_SERVICE_PORT", "8085")
	redisAdr := getenv("REDIS_ADDR", "redis:6379")
	redisPwd := getenv("REDIS_PASSWORD", "")
	redisDB := atoi(getenv("REDIS_DB", "0"))
	ttlSec := atoi(getenv("CACHE_TTL_SEC", "300"))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	svc := service.New(redisAdr, redisPwd, redisDB, ttlSec)
	hnd := handlers.New(svc)

	r := mux.NewRouter()
	r.HandleFunc("/cache/conversation/{u1}/{u2}", hnd.SetConv).Methods(http.MethodPost)
	r.HandleFunc("/cache/conversation/{u1}/{u2}", hnd.GetConv).Methods(http.MethodGet)
	r.HandleFunc("/cache/conversation/{u1}/{u2}", hnd.DelConv).Methods(http.MethodDelete)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info("cache-service started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down")
	_ = srv.Close()
}

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"message-service/internal/broker"
	"message-service/internal/handlers"
	"message-service/internal/repository"
	"message-service/internal/service"

	"github.com/gorilla/mux"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	port := getenv("MESSAGE_SERVICE_PORT", "8083")
	dbURL := getenv("MESSAGE_DB_URL", "postgres://root:root@localhost:5432/main_db?sslmode=disable")
	rmqURL := getenv("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/")
	rmqExchange := getenv("RABBIT_EXCHANGE", "msg.events")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := repository.NewDB(dbURL)
	if err != nil {
		logger.Error("db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	var br *broker.Broker
	if b, err := broker.New(rmqURL, rmqExchange); err == nil {
		br = b
		logger.Info("rabbitmq connected")
	} else {
		logger.Warn("rabbitmq unavailable; продолжение без очереди", "err", err)
	}

	msgRepo := repository.NewMessageRepository(db)
	srv := service.NewMessageService(msgRepo, br)
	hnd := handlers.NewMessageHandlers(srv, logger)

	r := mux.NewRouter()
	r.HandleFunc("/messages/create", hnd.CreateMessageHandler).Methods(http.MethodPost)
	r.HandleFunc("/messages/get/{id}", hnd.GetMessageHandler).Methods(http.MethodGet)
	r.HandleFunc("/messages/delete/{id}", hnd.DeleteMessageHandler).Methods(http.MethodDelete)
	r.HandleFunc("/messages/conversation/{partner}", hnd.ConversationHandler).Methods(http.MethodGet)
	r.HandleFunc("/messages/dialogs", hnd.DialogsHandler).Methods(http.MethodGet)

	httpSrv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("message-service started", "port", port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	if br != nil {
		br.Close()
	}
	logger.Info("server stopped")
}

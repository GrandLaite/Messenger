package main

import (
	"log"
	"net/http"
	"os"

	"user-service/internal/handlers"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := getenv("USER_SERVICE_PORT", "8082")
	dbURL := getenv("USER_DB_URL", "postgres://root:root@localhost:5432/main_db?sslmode=disable")

	db, err := repository.NewDB(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	srv := service.NewUserService(repo)
	hnd := handlers.NewUserHandlers(srv)

	r := mux.NewRouter()
	r.HandleFunc("/users/create", hnd.CreateUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/users/checkpassword", hnd.CheckPasswordHandler).Methods(http.MethodPost)
	r.HandleFunc("/users/search/{nickname}", hnd.SearchUserHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/info/{nickname}", hnd.InfoUserHandler).Methods(http.MethodGet) // 🆕

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	log.Printf("User service on port %s", port)
	log.Fatal(server.ListenAndServe())
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

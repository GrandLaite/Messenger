package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"auth-service/internal/service"
	"log/slog"
)

type AuthHandlers struct {
	srv    *service.AuthService
	logger *slog.Logger
}

func NewAuthHandlers(s *service.AuthService, lg *slog.Logger) *AuthHandlers {
	return &AuthHandlers{srv: s, logger: lg}
}

func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	token, err := h.srv.LoginUser(r.Context(), req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Неверные имя пользователя или пароль")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	userSvcURL := getenv("USER_SERVICE_URL", "http://localhost:8082") + "/users/create"

	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Post(userSvcURL, "application/json", mustMarshal(req))
	if err != nil {
		respondError(w, http.StatusBadGateway, "Сервис пользователей недоступен")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respondError(w, resp.StatusCode, "Ошибка при регистрации пользователя")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"message": "Пользователь успешно зарегистрирован"})
}

func respondError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func mustMarshal(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

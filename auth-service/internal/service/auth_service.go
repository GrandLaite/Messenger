package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour
const bcryptCost = bcrypt.DefaultCost

type AuthService struct {
	secret string
	client *http.Client
	logger *slog.Logger
}

func NewAuthService(sec string, lg *slog.Logger) *AuthService {
	return &AuthService{
		secret: sec,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: lg,
	}
}

func (s *AuthService) GenerateToken(nickname, role string) (string, error) {
	claims := jwt.MapClaims{
		"nickname": nickname,
		"role":     role,
		"exp":      time.Now().Add(tokenTTL).Unix(),
		"iat":      time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.secret))
}

func (s *AuthService) LoginUser(ctx context.Context, username, password string) (string, error) {
	url := getenv("USER_SERVICE_URL", "http://localhost:8082") + "/users/checkpassword"

	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("ошибка проверки пользователя")
	}

	var res struct {
		Role     string `json:"role"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return s.GenerateToken(res.Nickname, res.Role)
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	return string(b), err
}

func CheckPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

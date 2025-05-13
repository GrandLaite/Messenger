package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type userInfo struct {
	Email string `json:"email"`
}

func getRecipientEmail(nickname string) (string, error) {
	base := os.Getenv("USER_SERVICE_URL")
	if base == "" {
		base = "http://user-service:8082"
	}
	url := base + "/users/info/" + nickname

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("user-service responded " + resp.Status)
	}

	var u userInfo
	if err = json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Email, nil
}

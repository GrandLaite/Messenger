package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

var (
	cacheBase = func() string {
		if v := os.Getenv("CACHE_SERVICE_URL"); v != "" {
			return v
		}
		return "http://cache-service:8085"
	}()
	httpCli = &http.Client{Timeout: 3 * time.Second}
)

func cacheURL(u1, u2 string) string {
	return cacheBase + "/cache/conversation/" + u1 + "/" + u2
}

// попытка чтения из Redis; возвращает true, если удалось
func tryCacheGet(ctx context.Context, u1, u2 string, dst any) bool {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cacheURL(u1, u2), nil)
	resp, err := httpCli.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dst) == nil
}

// асинхронная запись в кэш
func cacheSetAsync(u1, u2 string, data any) {
	go func() {
		b, _ := json.Marshal(data)
		_, _ = httpCli.Post(cacheURL(u1, u2), "application/json", bytes.NewBuffer(b))
	}()
}

// асинхронное удаление кэша
func cacheDelAsync(u1, u2 string) {
	go func() {
		req, _ := http.NewRequest(http.MethodDelete, cacheURL(u1, u2), nil)
		_, _ = httpCli.Do(req)
	}()
}

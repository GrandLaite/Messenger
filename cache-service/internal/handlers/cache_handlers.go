package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cache-service/internal/service"

	"github.com/gorilla/mux"
)

type Handlers struct{ srv *service.CacheService }

func New(s *service.CacheService) *Handlers { return &Handlers{srv: s} }

func respond(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func (h *Handlers) SetConv(w http.ResponseWriter, r *http.Request) {
	var data any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	u1, u2 := mux.Vars(r)["u1"], mux.Vars(r)["u2"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := h.srv.SetConversation(ctx, u1, u2, data); err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "cached"})
}

func (h *Handlers) GetConv(w http.ResponseWriter, r *http.Request) {
	u1, u2 := mux.Vars(r)["u1"], mux.Vars(r)["u2"]
	var out any
	ok, err := h.srv.GetConversation(r.Context(), u1, u2, &out)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !ok {
		respond(w, http.StatusNotFound, map[string]string{"error": "not cached"})
		return
	}
	respond(w, http.StatusOK, out)
}

func (h *Handlers) DelConv(w http.ResponseWriter, r *http.Request) {
	u1, u2 := mux.Vars(r)["u1"], mux.Vars(r)["u2"]
	if err := h.srv.DeleteConversation(r.Context(), u1, u2); err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respond(w, http.StatusOK, nil)
}

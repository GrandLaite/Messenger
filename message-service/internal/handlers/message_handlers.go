package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"message-service/internal/service"

	"log/slog"

	"github.com/gorilla/mux"
)

type MessageHandlers struct {
	srv    *service.MessageService
	logger *slog.Logger
}

func NewMessageHandlers(s *service.MessageService, lg *slog.Logger) *MessageHandlers {
	return &MessageHandlers{srv: s, logger: lg}
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

func (h *MessageHandlers) CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	sender := r.Header.Get("X-User-Name")
	var req struct {
		RecipientNickname string `json:"recipient_nickname"`
		Content           string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	msg, err := h.srv.Create(sender, req.RecipientNickname, req.Content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось создать сообщение")
		return
	}
	respondJSON(w, http.StatusOK, msg)
}

func (h *MessageHandlers) GetMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный идентификатор сообщения")
		return
	}
	msg, err := h.srv.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Сообщение не найдено")
		return
	}
	respondJSON(w, http.StatusOK, msg)
}

func (h *MessageHandlers) DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	sender := r.Header.Get("X-User-Name")
	userRole := r.Header.Get("X-User-Role")
	if userRole != "premium" {
		respondError(w, http.StatusForbidden, "Удаление доступно только премиум-пользователям")
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный идентификатор сообщения")
		return
	}
	if err := h.srv.Delete(id, sender); err != nil {
		respondError(w, http.StatusForbidden, "Нет прав на удаление")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MessageHandlers) ConversationHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("X-User-Name")
	partner := mux.Vars(r)["partner"]
	msgs, err := h.srv.GetConversation(user, partner)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось получить переписку")
		return
	}
	respondJSON(w, http.StatusOK, msgs)
}

func (h *MessageHandlers) DialogsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("X-User-Name")
	dialogs, err := h.srv.GetDialogs(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось получить список диалогов")
		return
	}
	respondJSON(w, http.StatusOK, dialogs)
}

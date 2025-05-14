package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"storage-service/internal/service"

	"github.com/gorilla/mux"
)

type StorageHandlers struct {
	svc    *service.StorageService
	logger *slog.Logger
}

func NewStorageHandlers(s *service.StorageService, lg *slog.Logger) *StorageHandlers {
	return &StorageHandlers{svc: s, logger: lg}
}

func (h *StorageHandlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		h.logger.Error("Invalid content type", "content-type", r.Header.Get("Content-Type"))
		http.Error(w, `{"error":"Content-Type must be multipart/form-data"}`, http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		h.logger.Error("Failed to parse form", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File too large (max 20MB)"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing 'file' in form-data"})
		return
	}
	defer file.Close()

	objectName, err := h.svc.UploadFile(r.Context(), "docs", file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("Upload failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "File upload failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"filename": objectName,
		"status":   "success",
		"size":     fmt.Sprintf("%d bytes", header.Size),
	})
}

func (h *StorageHandlers) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	obj, err := h.svc.DownloadFile(r.Context(), "docs", filename)
	if err != nil {
		h.logger.Error("Download failed", "error", err, "filename", filename)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	objInfo, err := obj.Stat()
	if err != nil {
		h.logger.Error("Failed to get file info", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", objInfo.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(objInfo.Size, 10))

	_, err = io.Copy(w, obj)
	if err != nil {
		h.logger.Error("Failed to stream file", "error", err)
	}
}

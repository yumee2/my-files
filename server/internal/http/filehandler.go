package http

import (
	"context"
	"encoding/json"
	"errors"
	"file-uploader/internal/models"
	"file-uploader/internal/service/dto"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type FileServiceI interface {
	AddFile(ctx context.Context, fileDTO *dto.UploadDTO) (string, error)
	DownloadFile(ctx context.Context, id string, w io.Writer) (*models.File, func(io.Writer) error, error)
	GetFiles(ctx context.Context) ([]*models.File, error)
	DeleteFile(ctx context.Context, id string) error
}

type FileHandler struct {
	service FileServiceI
}

func NewFileHandler(service FileServiceI) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	files, err := h.service.GetFiles(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := r.PathValue("id")
	ctx := r.Context()
	file, streamFunc, err := h.service.DownloadFile(ctx, fileUUID, w)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))

	err = streamFunc(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := r.PathValue("id")
	ctx := r.Context()
	if err := h.service.DeleteFile(ctx, fileUUID); err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success!"))
}

func (h *FileHandler) AddFile(w http.ResponseWriter, r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "expected multipart/form-data", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			http.Error(w, "file field not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if part.FormName() != "file" {
			part.Close()
			continue
		}

		fileDTO := &dto.UploadDTO{
			Name: part.FileName(),
			Body: part,
		}

		id, err := h.service.AddFile(ctx, fileDTO)
		part.Close()
		if err != nil {
			if errors.Is(err, models.ErrFileAlreadyExists) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id": id,
		})
		return
	}
}

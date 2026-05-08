package service

import (
	"context"
	"errors"
	"file-uploader/internal/service/dto"
	storage "file-uploader/internal/storage/filesystem"
	"file-uploader/models"
	"io"
	"log"
	"mime"
	"path/filepath"

	"github.com/google/uuid"
)

type DataBaseI interface {
	AddFile(ctx context.Context, file *models.File) error
	GetFile(ctx context.Context, id string) (*models.File, error)
	GetFiles(ctx context.Context) ([]*models.File, error)
	DeleteFile(ctx context.Context, id string) error
}

type FileService struct {
	db DataBaseI
}

func NewFileService(db DataBaseI) *FileService {
	return &FileService{db: db}
}

func (fs *FileService) AddFile(ctx context.Context, fileDTO *dto.UploadDTO) (string, error) {
	fileUUID := uuid.NewString()
	size, err := storage.SaveFile(ctx, fileUUID, fileDTO.Body)
	if err != nil {
		log.Print("failed to save file on a disk: ", err)
		return "", err
	}

	file := &models.File{
		ID:           fileUUID,
		OriginalName: filepath.Base(fileDTO.Name),
		MimeType:     mime.TypeByExtension(filepath.Ext(fileDTO.Name)),
		Size:         size,
	}

	err = fs.db.AddFile(ctx, file)
	if err != nil {
		if errors.Is(err, models.ErrFileAlreadyExists) {
			return fileUUID, models.ErrFileAlreadyExists
		}
		storage.DeleteFile(file.ID)
		log.Print("failed to save file in the database: ", err)
		return "", err
	}

	return fileUUID, nil
}

func (fs *FileService) DownloadFile(ctx context.Context, id string, w io.Writer) (*models.File, func(io.Writer) error, error) {
	file, err := fs.db.GetFile(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			return nil, nil, models.ErrFileNotFound
		}
		log.Print("failed to get the file from the database: ", err)
		return nil, nil, err
	}

	streamFunc := func(w io.Writer) error {
		return storage.WriteFileTo(ctx, id, w)
	}
	return file, streamFunc, nil
}

func (fs *FileService) GetFiles(ctx context.Context) ([]*models.File, error) {
	files, err := fs.db.GetFiles(ctx)
	if err != nil {
		log.Print("failed to get all files from the database: ", err)
		return nil, err
	}
	return files, nil
}

func (fs *FileService) DeleteFile(ctx context.Context, id string) error {
	err := storage.DeleteFile(id)
	if err != nil {
		if errors.Is(err, models.ErrFileNotFound) {
			return models.ErrFileNotFound
		}
		log.Print("failed to delete the file from the database: ", err)
		return err
	}
	return fs.db.DeleteFile(ctx, id)
}

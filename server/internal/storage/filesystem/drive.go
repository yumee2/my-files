package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const chunkSizeKB = 500
const chunkSizeBytes = 500 * 1024

var rootDir, _ = os.Getwd()

func SaveFile(ctx context.Context, uuid string, data io.Reader) (int64, error) {
	dirPath := filepath.Join(rootDir, "data", "chunks", uuid)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return 0, fmt.Errorf("failed to create chunk directory: %w", err)
	}

	buf := make([]byte, chunkSizeBytes)
	var totalSize int64 = 0
	var chunkIndex = 0

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
		n, err := io.ReadFull(data, buf)
		if n > 0 {
			totalSize += int64(n)
			chunkPath := filepath.Join(dirPath, fmt.Sprintf("chunk_%06d", chunkIndex))
			if err := os.WriteFile(chunkPath, buf[:n], 0644); err != nil {
				return 0, fmt.Errorf("failed to write chunk: %w", err)
			}
			chunkIndex++
		}
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				break
			}
			return 0, fmt.Errorf("failed to read chunk: %w", err)
		}
	}
	return totalSize, nil

}

func WriteFileTo(ctx context.Context, fileUUID string, w io.Writer) error {
	dirPath := filepath.Join(rootDir, "data", "chunks", fileUUID)
	entities, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read chunk directory: %w", err)
	}

	buf := make([]byte, chunkSizeBytes)
	for _, entity := range entities {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		filePath := filepath.Join(dirPath, entity.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open chunk %s: %w", entity.Name(), err)
		}
		_, copyErr := io.CopyBuffer(w, file, buf)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("failed to write chunk %s: %w", entity.Name(), copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close chunk %s: %w", entity.Name(), closeErr)
		}
	}
	return nil

}

func DeleteFile(fileUUID string) error {
	dirPath := filepath.Join(rootDir, "data", "chunks", fileUUID)
	return os.RemoveAll(dirPath)
}

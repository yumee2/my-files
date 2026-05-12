package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"file-uploader/internal/config"
	"file-uploader/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func NewDBConnection() (*Repository, error) {
	dataDir := config.DataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "files.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS files (
            id            TEXT PRIMARY KEY,
            original_name TEXT,
            size          INTEGER,
            mime_type     TEXT,
            created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS auth (
        	id            INTEGER PRIMARY KEY CHECK (id = 1),
         	password_hash TEXT NOT NULL,
        	created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
        	updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS sessions (
        	id         TEXT PRIMARY KEY,
        	expires_at DATETIME NOT NULL,
        	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (d *Repository) AddFile(ctx context.Context, file *models.File) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO files (id, original_name, size, mime_type) VALUES (?, ?, ?, ?)",
		file.ID, file.OriginalName, file.Size, file.MimeType)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return models.ErrFileAlreadyExists
		}
		return fmt.Errorf("failed to insert file: %w", err)
	}

	return nil
}

func (d *Repository) GetFile(ctx context.Context, id string) (*models.File, error) {
	rows := d.db.QueryRowContext(ctx, "SELECT id, original_name, size, mime_type FROM files WHERE id = ?", id)
	var file models.File
	if err := rows.Scan(&file.ID, &file.OriginalName, &file.Size, &file.MimeType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return &file, nil
}

func (d *Repository) GetFiles(ctx context.Context) ([]*models.File, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, original_name, size, mime_type FROM files")
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	files := make([]*models.File, 0)
	for rows.Next() {
		var file models.File
		if err := rows.Scan(&file.ID, &file.OriginalName, &file.Size, &file.MimeType); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, &file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return files, nil
}

func (d *Repository) DeleteFile(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "delete from files where id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrFileNotFound
	}

	return nil
}

func (d *Repository) CreatePassword(password string) error {
	_, err := d.db.Exec("INSERT INTO auth (id, password_hash) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET password_hash = ?", password, password)
	if err != nil {
		return fmt.Errorf("failed to insert password: %w", err)
	}

	return nil
}

func (d *Repository) GetPassword(ctx context.Context) (string, error) {
	var password string
	err := d.db.QueryRowContext(ctx, "SELECT password_hash FROM auth WHERE id = 1").Scan(&password)
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	return password, nil
}

func (d *Repository) CreateSession(ctx context.Context, id string, expiresAt time.Time) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO sessions (id, expires_at) VALUES (?, ?)", id, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (d *Repository) GetSession(ctx context.Context, id string) (*models.Session, error) {
	var session models.Session
	err := d.db.QueryRowContext(ctx, "SELECT id, expires_at, created_at FROM sessions WHERE id = ?", id).Scan(&session.ID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (d *Repository) Close() error {
	return d.db.Close()
}

// Package filestorage содержит реализацию хранилища на основе файла.
package filestorage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
)

// FileStorage представляет файловое хранилище.
type FileStorage struct {
	filePath string
}

// New создает новое файловое хранилище.
func New(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

// Load загружает все записи из файла.
func (fs *FileStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.URLRecord{}, nil
		}

		return nil, err
	}

	if len(data) == 0 {
		return []model.URLRecord{}, nil
	}

	var records []model.URLRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// Append добавляет запись в файл.
func (fs *FileStorage) Append(ctx context.Context, record model.URLRecord) error {
	records, err := fs.Load(ctx)
	if err != nil {
		return err
	}

	records = append(records, record)

	return fs.save(records)
}

// AppendBatch добавляет несколько записей в файл.
func (fs *FileStorage) AppendBatch(ctx context.Context, newRecords []model.URLRecord) error {
	if len(newRecords) == 0 {
		return nil
	}

	records, err := fs.Load(ctx)
	if err != nil {
		return err
	}

	records = append(records, newRecords...)

	return fs.save(records)
}

// FindByOriginalURL находит короткий URL по оригинальному.
func (fs *FileStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	records, err := fs.Load(ctx)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if record.OriginalURL == originalURL {
			return record.ShortURL, nil
		}
	}

	return "", nil
}

// FindByShortURL находит оригинальный URL по короткому.
func (fs *FileStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	records, err := fs.Load(ctx)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if record.ShortURL == shortURL {
			if record.IsDeleted {
				return "", repository.ErrDeleted
			}
			return record.OriginalURL, nil
		}
	}

	return "", nil
}

// FindByUserID находит все URL пользователя.
func (fs *FileStorage) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	records, err := fs.Load(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.URLRecord, 0)
	for _, record := range records {
		if record.UserID == userID {
			result = append(result, record)
		}
	}

	return result, nil
}

// DeleteBatch удаляет несколько URL пакетно.
func (fs *FileStorage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	records, err := fs.Load(ctx)
	if err != nil {
		return err
	}

	urlsMap := make(map[string]bool)
	for _, url := range shortURLs {
		urlsMap[url] = true
	}

	for i := range records {
		if records[i].UserID == userID && urlsMap[records[i].ShortURL] {
			records[i].IsDeleted = true
		}
	}

	return fs.save(records)
}

func (fs *FileStorage) save(records []model.URLRecord) error {
	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.filePath, data, 0644)
}

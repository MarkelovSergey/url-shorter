package filestorage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MarkelovSergey/url-shorter/internal/model"
)

type Storage interface {
	Load() ([]model.URLRecord, error)
	Append(record model.URLRecord) error
}

type fileStorage struct {
	filePath string
}

func New(filePath string) Storage {
	return &fileStorage{
		filePath: filePath,
	}
}

func (fs *fileStorage) Load() ([]model.URLRecord, error) {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return []model.URLRecord{}, nil
	}

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
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

func (fs *fileStorage) Append(record model.URLRecord) error {
	records, err := fs.Load()
	if err != nil {
		return err
	}

	records = append(records, record)

	return fs.save(records)
}

func (fs *fileStorage) save(records []model.URLRecord) error {
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
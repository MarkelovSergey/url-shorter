package filestorage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorageLoad(t *testing.T) {
	tests := []struct {
		name            string
		setupFile       func(string) error
		expectedRecords []model.URLRecord
		expectedError   bool
	}{
		{
			name: "successful load with existing records",
			setupFile: func(path string) error {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				data := []byte(`[
  {
    "uuid": "uuid-1",
    "short_url": "abc123",
    "original_url": "https://practicum.yandex.ru"
  },
  {
    "uuid": "uuid-2",
    "short_url": "def456",
    "original_url": "https://example.com"
  }
]`)
				return os.WriteFile(path, data, 0644)
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
			},
			expectedError: false,
		},
		{
			name: "load from non-existent file returns empty slice",
			setupFile: func(path string) error {
				return nil
			},
			expectedRecords: []model.URLRecord{},
			expectedError:   false,
		},
		{
			name: "load from empty file returns empty slice",
			setupFile: func(path string) error {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				return os.WriteFile(path, []byte(""), 0644)
			},
			expectedRecords: []model.URLRecord{},
			expectedError:   false,
		},
		{
			name: "load with invalid JSON returns error",
			setupFile: func(path string) error {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				return os.WriteFile(path, []byte("invalid json"), 0644)
			},
			expectedRecords: nil,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFilePath := filepath.Join(tempDir, "test-storage.json")

			err := tt.setupFile(testFilePath)
			require.NoError(t, err)

			storage := New(testFilePath)

			records, err := storage.Load()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRecords, records)
			}
		})
	}
}

func TestFileStorageAppend(t *testing.T) {
	tests := []struct {
		name            string
		existingData    []model.URLRecord
		recordToAppend  model.URLRecord
		expectedRecords []model.URLRecord
	}{
		{
			name:         "append to empty storage",
			existingData: []model.URLRecord{},
			recordToAppend: model.URLRecord{
				UUID:        "uuid-1",
				ShortURL:    "abc123",
				OriginalURL: "https://practicum.yandex.ru",
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
		},
		{
			name: "append to existing records",
			existingData: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
			recordToAppend: model.URLRecord{
				UUID:        "uuid-2",
				ShortURL:    "def456",
				OriginalURL: "https://example.com",
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
			},
		},
		{
			name: "append multiple records",
			existingData: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
			},
			recordToAppend: model.URLRecord{
				UUID:        "uuid-3",
				ShortURL:    "ghi789",
				OriginalURL: "https://test.com",
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
				{
					UUID:        "uuid-3",
					ShortURL:    "ghi789",
					OriginalURL: "https://test.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFilePath := filepath.Join(tempDir, "test-storage.json")
			storage := New(testFilePath)

			if len(tt.existingData) > 0 {
				fs := storage.(*fileStorage)
				err := fs.save(tt.existingData)
				require.NoError(t, err)
			}

			err := storage.Append(tt.recordToAppend)
			assert.NoError(t, err)

			records, err := storage.Load()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRecords, records)

			_, err = os.Stat(testFilePath)
			assert.NoError(t, err)
		})
	}
}

func TestFileStorageAppendCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "subdir1", "subdir2", "test-storage.json")
	storage := New(testFilePath)

	record := model.URLRecord{
		UUID:        "uuid-1",
		ShortURL:    "abc123",
		OriginalURL: "https://practicum.yandex.ru",
	}

	err := storage.Append(record)
	assert.NoError(t, err)

	dir := filepath.Dir(testFilePath)
	_, err = os.Stat(dir)
	assert.NoError(t, err)

	records, err := storage.Load()
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, record, records[0])
}

func TestFileStorageAppendBatch(t *testing.T) {
	tests := []struct {
		name            string
		existingData    []model.URLRecord
		recordsToAppend []model.URLRecord
		expectedRecords []model.URLRecord
	}{
		{
			name:         "append batch to empty storage",
			existingData: []model.URLRecord{},
			recordsToAppend: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
			},
		},
		{
			name: "append batch to existing records",
			existingData: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
			recordsToAppend: []model.URLRecord{
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
				{
					UUID:        "uuid-3",
					ShortURL:    "ghi789",
					OriginalURL: "https://test.com",
				},
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
				{
					UUID:        "uuid-2",
					ShortURL:    "def456",
					OriginalURL: "https://example.com",
				},
				{
					UUID:        "uuid-3",
					ShortURL:    "ghi789",
					OriginalURL: "https://test.com",
				},
			},
		},
		{
			name: "append empty batch does nothing",
			existingData: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
			recordsToAppend: []model.URLRecord{},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
		},
		{
			name:         "append batch with single record",
			existingData: []model.URLRecord{},
			recordsToAppend: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
			expectedRecords: []model.URLRecord{
				{
					UUID:        "uuid-1",
					ShortURL:    "abc123",
					OriginalURL: "https://practicum.yandex.ru",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testFilePath := filepath.Join(tempDir, "test-storage.json")
			storage := New(testFilePath)

			if len(tt.existingData) > 0 {
				fs := storage.(*fileStorage)
				err := fs.save(tt.existingData)
				require.NoError(t, err)
			}

			err := storage.AppendBatch(tt.recordsToAppend)
			assert.NoError(t, err)

			records, err := storage.Load()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRecords, records)

			if len(tt.recordsToAppend) > 0 {
				_, err = os.Stat(testFilePath)
				assert.NoError(t, err)
			}
		})
	}
}

package audit

import (
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"
)

// FileObserver записывает события аудита в файл.
type FileObserver struct {
	file   *os.File
	mu     *sync.Mutex
	logger *zap.Logger
}

// NewFileObserver создает новый наблюдатель для записи событий в файл.
func NewFileObserver(filePath string, logger *zap.Logger) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileObserver{
		file:   file,
		mu:     &sync.Mutex{},
		logger: logger,
	}, nil
}

// OnEvent обрабатывает событие аудита и записывает его в файл.
func (fo *FileObserver) OnEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		fo.logger.Error("failed to marshal audit event", zap.Error(err))
		return err
	}

	fo.mu.Lock()
	defer fo.mu.Unlock()

	if _, err := fo.file.Write(append(data, '\n')); err != nil {
		fo.logger.Error("failed to write audit event to file", zap.Error(err))
		return err
	}

	return nil
}

// Close закрывает файл аудита.
func (fo *FileObserver) Close() error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	if fo.file != nil {
		return fo.file.Close()
	}

	return nil
}

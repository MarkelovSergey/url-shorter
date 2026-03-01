package handler

import (
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/usecase/urlcase"
	"go.uber.org/zap"
)

// handler содержит зависимости для обработки HTTP-запросов.
type handler struct {
	config     config.Config
	urlUseCase urlcase.URLUseCase
	logger     *zap.Logger
}

// New создает новый экземпляр обработчика с заданными зависимостями.
// Возвращает указатель на handler, который содержит методы для обработки HTTP-запросов.
func New(
	config config.Config,
	urlUseCase urlcase.URLUseCase,
	logger *zap.Logger,
) *handler {
	return &handler{config, urlUseCase, logger}
}

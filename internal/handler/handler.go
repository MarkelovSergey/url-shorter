package handler

import (
	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"go.uber.org/zap"
)

// handler содержит зависимости для обработки HTTP-запросов.
type handler struct {
	config            config.Config
	urlShorterService urlshorterservice.URLShorterService
	healthService     healthservice.HealthService
	logger            *zap.Logger
	auditPublisher    audit.Publisher
}

// New создает новый экземпляр обработчика с заданными зависимостями.
// Возвращает указатель на handler, который содержит методы для обработки HTTP-запросов.
func New(
	config config.Config,
	urlShorterService urlshorterservice.URLShorterService,
	healthService healthservice.HealthService,
	logger *zap.Logger,
	auditPublisher audit.Publisher,
) *handler {
	return &handler{config, urlShorterService, healthService, logger, auditPublisher}
}

package handler

import (
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"go.uber.org/zap"
)

type handler struct {
	config            config.Config
	urlShorterService urlshorterservice.URLShorterService
	healthService     healthservice.HealthService
	logger            *zap.Logger
}

func New(
	config config.Config,
	urlShorterService urlshorterservice.URLShorterService,
	healthService     healthservice.HealthService,
	logger *zap.Logger,
) *handler {
	return &handler{config, urlShorterService, healthService, logger}
}

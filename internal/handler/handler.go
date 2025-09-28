package handler

import (
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
)

type handler struct {
	config            config.Config
	urlShorterService urlshorterservice.URLShorterService
}

func New(
	config config.Config,
	urlShorterService urlshorterservice.URLShorterService,
) *handler {
	return &handler{config, urlShorterService}
}

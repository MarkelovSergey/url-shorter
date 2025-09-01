package handler

import "github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"

type handler struct {
	urlShorterService urlshorterservice.URLShorterService
}

func New(urlShorterService urlshorterservice.URLShorterService) *handler {
	return &handler{urlShorterService}
}

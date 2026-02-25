// Package model содержит модели данных приложения.
package model

import "github.com/golang-jwt/jwt/v5"

// UserClaims содержит данные пользователя для JWT-токена.
type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Request представляет запрос на создание короткой ссылки.
type Request struct {
	URL string `json:"url"`
}

// Response представляет ответ с короткой ссылкой.
type Response struct {
	Result string `json:"result"`
}

// BatchRequest представляет элемент батч-запроса
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse представляет элемент батч-ответа
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URLRecord представляет запись сокращённого URL для сохранения в файл
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// UserURLResponse представляет элемент ответа для получения URL пользователя
type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// StatsResponse содержит статистику сервиса.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
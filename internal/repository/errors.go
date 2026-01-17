// Package repository содержит репозитории для работы с данными.
package repository

import "errors"

// Ошибки репозитория.
var (
	// ErrShortCodeAlreadyExist - короткий код уже существует.
	ErrShortCodeAlreadyExist = errors.New("shortcode already exists")
	// ErrNotFound - URL не найден.
	ErrNotFound = errors.New("not found")
	// ErrURLAlreadyExists - оригинальный URL уже существует.
	ErrURLAlreadyExists = errors.New("original URL already exists")
	// ErrDeleted - URL был удален.
	ErrDeleted = errors.New("url has been deleted")
)

// Package service содержит бизнес-логику приложения.
package service

import "errors"

// Ошибки сервисного слоя.
var (
	// ErrGenerateShortCode - ошибка генерации короткого кода.
	ErrGenerateShortCode = errors.New("failed to generate unique short code")
	// ErrSaveShortCode - ошибка сохранения короткого кода.
	ErrSaveShortCode = errors.New("failed to save short code")
	// ErrFindShortCode - ошибка поиска URL.
	ErrFindShortCode = errors.New("failed to find URL for short code")
	// ErrURLConflict - URL уже был сокращен.
	ErrURLConflict = errors.New("URL already shortened")
	// ErrURLDeleted - URL был удален.
	ErrURLDeleted = errors.New("URL has been deleted")
)

package service

import "errors"

var (
	ErrGenerateShortCode = errors.New("failed to generate unique short code")
	ErrSaveShortCode     = errors.New("failed to save short code")
	ErrFindShortCode     = errors.New("failed to find URL for short code")
	ErrURLConflict       = errors.New("URL already shortened")
)

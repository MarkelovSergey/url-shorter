package repository

import "errors"

var (
	ErrShortCodeAlreadyExist = errors.New("shortcode already exists")
	ErrNotFound              = errors.New("not found")
	ErrURLAlreadyExists      = errors.New("original URL already exists")
	ErrDeleted               = errors.New("url has been deleted")
)

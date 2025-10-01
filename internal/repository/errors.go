package repository

import "errors"

var (
	ErrURLAlreadyExists  = errors.New("URL already exists")
	ErrNotFound          = errors.New("not found")
)

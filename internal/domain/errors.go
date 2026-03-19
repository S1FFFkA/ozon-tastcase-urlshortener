package domain

import "errors"

var (
	ErrInternal         = errors.New("internal server error")
	ErrNotFound         = errors.New("not found")
	ErrBadRequest       = errors.New("bad request")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrInvalidURL       = errors.New("invalid original url")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrConflict         = errors.New("conflict")
)

package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
	ErrIdNotFound  = errors.New("id not found")

	UserNotFound = errors.New("user not found")
)

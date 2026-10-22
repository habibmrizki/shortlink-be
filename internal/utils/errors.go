package utils

import "errors"

var (
	// Auth Errors
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized: invalid or expired token")

	// Link Errors
	ErrInvalidURL       = errors.New("invalid original URL format, must start with http:// or https://")
	ErrSlugTaken        = errors.New("Slug already taken")
	ErrSlugInvalidLen   = errors.New("custom slug must be between 3 and 50 characters")
	ErrSlugInvalidChars = errors.New("custom slug can only contain alphanumeric characters and hyphens")
	ErrSlugReserved     = errors.New("custom slug cannot use reserved words (api, login, register, dashboard)")
	ErrLinkNotFound     = errors.New("link not found")
)

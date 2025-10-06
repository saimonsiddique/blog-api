package domain

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailTaken           = errors.New("email already taken")
	ErrUsernameTaken        = errors.New("username already taken")
	ErrPostNotFound         = errors.New("post not found")
	ErrSlugTaken            = errors.New("slug already taken")
	ErrForbidden            = errors.New("forbidden")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidToken         = errors.New("invalid token")
	ErrConflict             = errors.New("conflict")
	ErrPostAlreadyPublished = errors.New("post already published")
	ErrInvalidStatusChange  = errors.New("invalid status change")
)

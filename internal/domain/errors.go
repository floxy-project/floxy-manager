package domain

import (
	"errors"
)

var (
	ErrEntityNotFound       = errors.New("entity not found")
	ErrEntityAlreadyExists  = errors.New("entity already exists")
	ErrInvalidToken         = errors.New("invalid token")
	ErrUsernameAlreadyInUse = errors.New("username already in use")
	ErrEmailAlreadyInUse    = errors.New("email already in use")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInactiveUser         = errors.New("inactive user")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalid2FACode       = errors.New("invalid 2FA code")
	ErrInvalidEmailCode     = errors.New("invalid email code")
	ErrTwoFARequired        = errors.New("2FA required")
	ErrTooMany2FAAttempts   = errors.New("too many 2FA attempts, try later")
)

type SkippableError struct {
	err error
}

func NewSkippableError(err error) *SkippableError {
	return &SkippableError{err: err}
}

func (e *SkippableError) Error() string {
	return e.err.Error()
}

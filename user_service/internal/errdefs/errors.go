package errdefs

import "errors"

var (
	ErrAlreadyExists    = errors.New("user already exists")
	ValidationErr       = errors.New("validation error")
	AuthenticationErr   = errors.New("authentication error")
	ErrNotFound         = errors.New("user not found")
	ErrPermissionDenied = errors.New("permission denied")
)

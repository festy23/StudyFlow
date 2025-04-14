package errdefs

import "errors"

var (
	ErrAlreadyExists    = errors.New("already exists")
	ValidationErr       = errors.New("validation error")
	AuthenticationErr   = errors.New("authentication error")
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
)

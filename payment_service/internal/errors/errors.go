package errdefs

import "errors"

var (
	ErrPermissionDenied = errors.New("permission was denied")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrNotFound         = errors.New("not found")
	ErrInvalidPayment   = errors.New("invalid payment")
	ErrAlreadyExists    = errors.New("user already exists")
)

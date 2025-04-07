package errdefs

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

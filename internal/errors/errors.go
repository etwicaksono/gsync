package errors

import "errors"

var (
	ErrFailedParseAuthData = errors.New(`failed to parse auth data`)
	ErrLogLevelInvalid     = errors.New(`invalid log level`)
)

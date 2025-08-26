package errors

import "errors"

var (
	ErrLogLevelInvalid = errors.New(`invalid log level`)
)

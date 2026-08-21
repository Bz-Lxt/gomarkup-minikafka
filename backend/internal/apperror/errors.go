package apperror

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalid       = errors.New("invalid")
	ErrConflict      = errors.New("conflict")
	ErrUnavailable   = errors.New("unavailable")
)

func Wrap(err error, format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{err}, args...)...)
}

func IsNotFound(err error) bool      { return errors.Is(err, ErrNotFound) }
func IsAlreadyExists(err error) bool { return errors.Is(err, ErrAlreadyExists) }
func IsInvalid(err error) bool       { return errors.Is(err, ErrInvalid) }

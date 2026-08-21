package broker

import "minikafka/internal/apperror"

var (
	ErrNotFound      = apperror.ErrNotFound
	ErrAlreadyExists = apperror.ErrAlreadyExists
	ErrInvalid       = apperror.ErrInvalid
)

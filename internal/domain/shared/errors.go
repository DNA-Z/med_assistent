package shared

import "errors"

var (
	ErrInvalidID         = errors.New("invalid id")
	ErrInvalidDate       = errors.New("invalid date")
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrInvalidValue      = errors.New("invalid value")
	ErrEmptyValue        = errors.New("value cannot be empty")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNotFound          = errors.New("not found")
)

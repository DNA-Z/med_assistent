package ports

import "errors"

var (
	ErrExaminationNotFound = errors.New("examination not found")
	ErrDoctorNotFound      = errors.New("doctor not found")
	ErrInvalidCommand      = errors.New("invalid command")
	ErrEmptyQuestion       = errors.New("question cannot be empty")
	ErrEmptyKeyword        = errors.New("keyword cannot be empty")
	ErrFileRequired        = errors.New("file or transcript is required")
	ErrPatientRequired     = errors.New("patient is required")
	ErrProcessingQueueFull = errors.New("processing queue is full")
)

package ports

import (
	"context"
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

type ProcessingTask struct {
	ID            uuid.UUID
	ExaminationID uuid.UUID
	Status        value_object.ProcessingStatus
	ErrorMessage  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProcessingRepository interface {
	Create(
		ctx context.Context,
		task *ProcessingTask,
	) error

	GetByExaminationID(
		ctx context.Context,
		examinationID uuid.UUID,
	) (*ProcessingTask, error)

	UpdateStatus(
		ctx context.Context,
		taskID uuid.UUID,
		status value_object.ProcessingStatus,
		errorMessage string,
	) error
}

package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type AuthService interface {
	Start(ctx context.Context, doctorID int64) error
}

type ExaminationCommandHandler interface {
	Load(ctx context.Context, cmd LoadExaminationCommand) (uuid.UUID, error)
	Retry(ctx context.Context, cmd RetryExaminationCommand) error
	Delete(ctx context.Context, cmd DeleteExaminationCommand) error
}

type LoadExaminationCommand struct {
	DoctorID  int64
	PatientID uuid.UUID

	FileName string
	File     io.ReadCloser

	Transcript *string
}

type RetryExaminationCommand struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

type DeleteExaminationCommand struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

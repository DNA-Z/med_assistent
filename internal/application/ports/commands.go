package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// AuthService регистрирует врача при первом обращении.
type AuthService interface {
	Start(ctx context.Context, doctorID int64) error
}

// ExaminationCommandHandler задаёт входной порт команд обследования.
type ExaminationCommandHandler interface {
	Load(ctx context.Context, cmd LoadExaminationCommand) (uuid.UUID, error)
	Retry(ctx context.Context, cmd RetryExaminationCommand) error
	Delete(ctx context.Context, cmd DeleteExaminationCommand) error
}

// LoadExaminationCommand содержит данные для создания и обработки обследования.
type LoadExaminationCommand struct {
	DoctorID  int64
	PatientID uuid.UUID

	FileName string
	File     io.ReadCloser

	Transcript *string
}

// RetryExaminationCommand содержит данные повторного запуска обработки.
type RetryExaminationCommand struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

// DeleteExaminationCommand содержит данные удаления обследования владельцем.
type DeleteExaminationCommand struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

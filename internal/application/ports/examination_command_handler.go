package ports

import (
	"context"

	"github.com/google/uuid"
)

type AuthService interface {
	Start(ctx context.Context, telegramUserID int64) error
}

type ExaminationCommandHandler interface {
	Load(ctx context.Context, cmd LoadExaminationCommand) (uuid.UUID, error)
	Retry(ctx context.Context, cmd RetryExaminationCommand) error
	Delete(ctx context.Context, cmd DeleteExaminationCommand) error
}

type LoadExaminationCommand struct {
	DoctorTelegramID int64
	FileName         string
	FilePath         string
}

type RetryExaminationCommand struct {
	DoctorTelegramID int64
	ExaminationID    uuid.UUID
}

type DeleteExaminationCommand struct {
	DoctorTelegramID int64
	ExaminationID    uuid.UUID
}

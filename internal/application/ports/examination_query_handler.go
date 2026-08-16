package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ExaminationQueryHandler interface {
	List(ctx context.Context, query ListExaminationsQuery) ([]ExaminationReadModel, error)
	Status(ctx context.Context, query GetExaminationStatusQuery) (*ExaminationStatusReadModel, error)
	Get(ctx context.Context, query GetExaminationQuery) (*ExaminationReadModel, error)
	Find(ctx context.Context, query FindExaminationsQuery) ([]ExaminationReadModel, error)
	Chat(ctx context.Context, query ChatQuery) (string, error)
}

type ListExaminationsQuery struct {
	DoctorTelegramID int64
}

type GetExaminationStatusQuery struct {
	DoctorTelegramID int64
	ExaminationID    uuid.UUID
}

type GetExaminationQuery struct {
	DoctorTelegramID int64
	ExaminationID    uuid.UUID
}

type FindExaminationsQuery struct {
	DoctorTelegramID int64
	Keyword          string
}

type ChatQuery struct {
	DoctorTelegramID int64
	Text             string
}

type ExaminationReadModel struct {
	ID              uuid.UUID
	PatientName     string
	PatientLastName string
	ExaminationDate time.Time
	Status          string
	Summary         string
}

type ExaminationStatusReadModel struct {
	ID        uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Error     string
}

package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ExaminationQueryHandler interface {
	List(ctx context.Context, query ListExaminationsQuery) ([]ExaminationDTO, error)
	Status(ctx context.Context, query GetExaminationStatusQuery) (*ExaminationStatusDTO, error)
	Get(ctx context.Context, query GetExaminationQuery) (*ExaminationDTO, error)
	Find(ctx context.Context, query FindExaminationsQuery) ([]ExaminationDTO, error)
	Chat(ctx context.Context, query ChatQuery) (string, error)
}

type ListExaminationsQuery struct {
	DoctorID int64
}

type GetExaminationStatusQuery struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

type GetExaminationQuery struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

type FindExaminationsQuery struct {
	DoctorID int64
	Keyword  string
}

type ChatQuery struct {
	DoctorID      int64
	ExaminationID *uuid.UUID
	Question      string
}

type ExaminationDTO struct {
	ID              uuid.UUID
	DoctorID        int64
	PatientID       uuid.UUID
	ExaminationDate time.Time

	Status      string
	Transcript  string
	Summary     string
	Diagnosis   string
	ErrorReason string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExaminationStatusDTO struct {
	ID        uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Error     string
}

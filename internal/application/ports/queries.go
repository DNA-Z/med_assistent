package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ExaminationQueryHandler задаёт входной порт запросов к read model.
type ExaminationQueryHandler interface {
	List(ctx context.Context, query ListExaminationsQuery) ([]ExaminationDTO, error)
	Status(ctx context.Context, query GetExaminationStatusQuery) (*ExaminationStatusDTO, error)
	Get(ctx context.Context, query GetExaminationQuery) (*ExaminationDTO, error)
	Find(ctx context.Context, query FindExaminationsQuery) ([]ExaminationDTO, error)
	Chat(ctx context.Context, query ChatQuery) (string, error)
}

// ListExaminationsQuery запрашивает обследования текущего врача.
type ListExaminationsQuery struct {
	DoctorID int64
}

// GetExaminationStatusQuery запрашивает статус принадлежащего врачу обследования.
type GetExaminationStatusQuery struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

// GetExaminationQuery запрашивает полные данные обследования.
type GetExaminationQuery struct {
	DoctorID      int64
	ExaminationID uuid.UUID
}

// FindExaminationsQuery задаёт полнотекстовый запрос в пределах данных врача.
type FindExaminationsQuery struct {
	DoctorID int64
	Keyword  string
}

// ChatQuery содержит вопрос и необязательное ограничение одним обследованием.
type ChatQuery struct {
	DoctorID      int64
	ExaminationID *uuid.UUID
	Question      string
}

// ExaminationDTO представляет денормализованную модель чтения обследования.
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

// ExaminationStatusDTO представляет состояние фоновой обработки.
type ExaminationStatusDTO struct {
	ID        uuid.UUID
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Error     string
}

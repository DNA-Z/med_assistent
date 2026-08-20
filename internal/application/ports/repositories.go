package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ExaminationWriteRepository interface {
	CreateExamination(
		ctx context.Context,
		examination ExaminationWriteModel,
		job ProcessingJobWriteModel,
	) error

	StartProcessing(
		ctx context.Context,
		examinationID uuid.UUID,
		jobID uuid.UUID,
		startedAt time.Time,
		attempt int,
	) error

	SaveTranscript(
		ctx context.Context,
		examinationID uuid.UUID,
		jobID uuid.UUID,
		transcript string,
		updatedAt time.Time,
	) error

	SaveSummary(
		ctx context.Context,
		examinationID uuid.UUID,
		summary string,
		updatedAt time.Time,
	) error

	CompleteProcessing(
		ctx context.Context,
		examinationID uuid.UUID,
		jobID uuid.UUID,
		updatedAt time.Time,
	) error

	FailProcessing(
		ctx context.Context,
		examinationID uuid.UUID,
		jobID uuid.UUID,
		errText string,
		updatedAt time.Time,
	) error

	RetryProcessing(
		ctx context.Context,
		doctorID int64,
		examinationID uuid.UUID,
		jobID uuid.UUID,
		updatedAt time.Time,
	) (uuid.UUID, string, error)

	DeleteExamination(
		ctx context.Context,
		doctorID int64,
		examinationID uuid.UUID,
	) error
}

type DoctorWriteRepository interface {
	Exists(ctx context.Context, doctorID int64) (bool, error)
	Create(ctx context.Context, doctorID int64) error
}

type ExaminationReadRepository interface {
	List(ctx context.Context, doctorID int64) ([]ExaminationDTO, error)

	Get(
		ctx context.Context,
		doctorID int64,
		examinationID uuid.UUID,
	) (*ExaminationDTO, error)

	Status(
		ctx context.Context,
		doctorID int64,
		examinationID uuid.UUID,
	) (*ExaminationStatusDTO, error)

	Find(
		ctx context.Context,
		doctorID int64,
		keyword string,
	) ([]ExaminationDTO, error)

	ChatContext(
		ctx context.Context,
		doctorID int64,
		examinationID *uuid.UUID,
	) ([]ChatContextItem, error)
}

type ChatContextItem struct {
	ExaminationID uuid.UUID
	Transcript    string
	Summary       string
}

type ExaminationWriteModel struct {
	ID              uuid.UUID
	DoctorID        int64
	PatientID       uuid.UUID
	ExaminationDate time.Time
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProcessingJobWriteModel struct {
	ID            uuid.UUID
	ExaminationID uuid.UUID
	Status        string
	Attempt       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

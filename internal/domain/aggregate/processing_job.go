package aggregate

import (
	"errors"
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

var (
	ErrInvalidProcessingJobID  = errors.New("некорректный идентификатор задания обработки")
	ErrInvalidProcessingExamID = errors.New("некорректный идентификатор обследования в задании")
	ErrInvalidAttemptCount     = errors.New("некорректное количество попыток")
	ErrJobAlreadyStarted       = errors.New("задание обработки уже запущено")
	ErrJobAlreadyCompleted     = errors.New("задание обработки уже завершено")
	ErrJobCannotRetry          = errors.New("задание обработки нельзя повторить")
	ErrJobNotProcessing        = errors.New("задание не находится в состоянии обработки")
	ErrProcessingErrorRequired = errors.New("требуется ошибка обработки")
)

type ProcessingJob struct {
	id            uuid.UUID
	examinationID uuid.UUID

	status  value_object.ProcessingStatus
	attempt int

	errorReason string

	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	updatedAt   time.Time
}

func NewProcessingJob(
	id uuid.UUID,
	examinationID uuid.UUID,
) (*ProcessingJob, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidProcessingJobID
	}

	if examinationID == uuid.Nil {
		return nil, ErrInvalidProcessingExamID
	}

	now := time.Now().UTC()

	return &ProcessingJob{
		id:            id,
		examinationID: examinationID,
		status:        value_object.ProcessingCreated,
		attempt:       0,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func (j *ProcessingJob) Start() error {
	if j.status != value_object.ProcessingCreated &&
		j.status != value_object.ProcessingFailed {
		return ErrJobAlreadyStarted
	}

	now := time.Now().UTC()

	j.status = value_object.ProcessingProcessing
	j.attempt++
	j.errorReason = ""
	j.startedAt = &now
	j.completedAt = nil
	j.touch()

	return nil
}

func (j *ProcessingJob) Complete() error {
	if j.status != value_object.ProcessingProcessing {
		return ErrJobNotProcessing
	}

	now := time.Now().UTC()

	j.status = value_object.ProcessingCompleted
	j.completedAt = &now

	j.touch()

	return nil
}

func (j *ProcessingJob) Fail(err error) error {
	if j.status != value_object.ProcessingProcessing {
		return ErrJobNotProcessing
	}

	if err == nil {
		return ErrProcessingErrorRequired
	}

	j.status = value_object.ProcessingFailed
	j.errorReason = err.Error()

	j.touch()

	return nil
}

func (j *ProcessingJob) Retry() error {
	if j.status != value_object.ProcessingFailed {
		return ErrJobCannotRetry
	}

	j.status = value_object.ProcessingCreated
	j.errorReason = ""
	j.completedAt = nil

	j.touch()

	return nil
}

func (j ProcessingJob) ID() uuid.UUID {
	return j.id
}

func (j ProcessingJob) ExaminationID() uuid.UUID {
	return j.examinationID
}

func (j ProcessingJob) Status() value_object.ProcessingStatus {
	return j.status
}

func (j ProcessingJob) Attempt() int {
	return j.attempt
}

func (j ProcessingJob) ErrorReason() string {
	return j.errorReason
}

func (j ProcessingJob) CreatedAt() time.Time {
	return j.createdAt
}

func (j ProcessingJob) StartedAt() *time.Time {
	return j.startedAt
}

func (j ProcessingJob) CompletedAt() *time.Time {
	return j.completedAt
}

func (j ProcessingJob) UpdatedAt() time.Time {
	return j.updatedAt
}

func (j *ProcessingJob) touch() {
	j.updatedAt = time.Now().UTC()
}

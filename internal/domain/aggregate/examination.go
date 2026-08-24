package aggregate

import (
	"errors"
	"strings"
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
)

var (
	ErrInvalidExaminationID   = errors.New("некорректный идентификатор обследования")
	ErrInvalidDoctorID        = errors.New("некорректный идентификатор врача")
	ErrInvalidPatientID       = errors.New("некорректный идентификатор пациента")
	ErrInvalidExaminationDate = errors.New("некорректная дата обследования")

	ErrTranscriptAlreadyExists = errors.New("транскрипция уже существует")
	ErrSummaryAlreadyExists    = errors.New("выжимка уже существует")
	ErrDiagnosisAlreadyExists  = errors.New("диагноз уже существует")

	ErrTranscriptRequired = errors.New("требуется транскрипция")
	ErrSummaryRequired    = errors.New("требуется выжимка")
	ErrDiagnosisRequired  = errors.New("требуется диагноз")

	ErrExaminationNotProcessing = errors.New(
		"обследование не находится в состоянии обработки",
	)

	ErrExaminationNotReady = errors.New(
		"обследование не готово к завершению",
	)
)

type Examination struct {
	id              uuid.UUID
	examinationDate time.Time

	doctorID  uuid.UUID
	patientID uuid.UUID

	transcript *value_object.Transcript
	summary    *value_object.BriefSummary
	diagnosis  *value_object.Diagnosis

	status      value_object.ExaminationStatus
	errorReason string

	createdAt time.Time
	updatedAt time.Time
}

func NewExamination(
	id uuid.UUID,
	examinationDate time.Time,
	doctorID uuid.UUID,
	patientID uuid.UUID,
) (*Examination, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidExaminationID
	}

	if doctorID == uuid.Nil {
		return nil, ErrInvalidDoctorID
	}

	if patientID == uuid.Nil {
		return nil, ErrInvalidPatientID
	}

	if examinationDate.IsZero() {
		return nil, ErrInvalidExaminationDate
	}

	now := time.Now().UTC()

	return &Examination{
		id:              id,
		examinationDate: examinationDate,
		doctorID:        doctorID,
		patientID:       patientID,
		status:          value_object.ExaminationCreated,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func (e *Examination) StartProcessing() error {
	if e.status != value_object.ExaminationCreated &&
		e.status != value_object.ExaminationFailed {
		return errors.New("обследование нельзя запустить")
	}

	e.status = value_object.ExaminationProcessing
	e.errorReason = ""
	e.touch()

	return nil
}

func (e *Examination) SetTranscript(
	transcript value_object.Transcript,
) error {

	if e.status != value_object.ExaminationProcessing {
		return ErrExaminationNotProcessing
	}

	if transcript.IsEmpty() {
		return ErrTranscriptRequired
	}

	if e.transcript != nil {
		return ErrTranscriptAlreadyExists
	}

	e.transcript = &transcript
	e.status = value_object.ExaminationTranscribed

	e.touch()

	return nil
}

func (e *Examination) SetSummary(
	summary value_object.BriefSummary,
) error {

	if e.status != value_object.ExaminationTranscribed {
		return errors.New(
			"выжимку можно добавить только после транскрипции",
		)
	}

	if summary.IsEmpty() {
		return ErrSummaryRequired
	}

	if e.summary != nil {
		return ErrSummaryAlreadyExists
	}

	e.summary = &summary
	e.status = value_object.ExaminationSummarized

	e.touch()

	return nil
}

func (e *Examination) AssignDiagnosis(
	diagnosis value_object.Diagnosis,
) error {

	if e.status != value_object.ExaminationSummarized &&
		e.status != value_object.ExaminationCompleted {
		return errors.New(
			"диагноз можно назначить только после создания выжимки",
		)
	}

	if e.diagnosis != nil {
		return ErrDiagnosisAlreadyExists
	}

	e.diagnosis = &diagnosis

	e.touch()

	return nil
}

func (e *Examination) Complete() error {
	if e.transcript == nil {
		return ErrTranscriptRequired
	}

	if e.summary == nil {
		return ErrSummaryRequired
	}

	if e.diagnosis == nil {
		return ErrDiagnosisRequired
	}

	if e.status != value_object.ExaminationSummarized {
		return ErrExaminationNotReady
	}

	e.status = value_object.ExaminationCompleted
	e.touch()

	return nil
}

func (e *Examination) Fail(reason error) error {
	if reason == nil {
		return errors.New("требуется причина ошибки")
	}

	if e.status == value_object.ExaminationCompleted {
		return errors.New(
			"завершённое обследование нельзя перевести в состояние ошибки",
		)
	}

	e.status = value_object.ExaminationFailed
	e.errorReason = strings.TrimSpace(reason.Error())

	e.touch()

	return nil
}

func (e *Examination) ID() uuid.UUID {
	return e.id
}

func (e *Examination) ExaminationDate() time.Time {
	return e.examinationDate
}

func (e *Examination) DoctorID() uuid.UUID {
	return e.doctorID
}

func (e *Examination) PatientID() uuid.UUID {
	return e.patientID
}

func (e *Examination) Transcript() (value_object.Transcript, bool) {
	if e.transcript == nil {
		return value_object.Transcript{}, false
	}

	return *e.transcript, true
}

func (e *Examination) Summary() (value_object.BriefSummary, bool) {
	if e.summary == nil {
		return value_object.BriefSummary{}, false
	}

	return *e.summary, true
}

func (e *Examination) Diagnosis() (value_object.Diagnosis, bool) {
	if e.diagnosis == nil {
		return value_object.Diagnosis{}, false
	}

	return *e.diagnosis, true
}

func (e *Examination) Status() value_object.ExaminationStatus {
	return e.status
}

func (e *Examination) ErrorReason() string {
	return e.errorReason
}

func (e *Examination) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Examination) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *Examination) touch() {
	e.updatedAt = time.Now().UTC()
}

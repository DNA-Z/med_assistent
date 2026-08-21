package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres/sqlqueries"
)

// ExaminationWriteRepository сохраняет write model и outbox-события в PostgreSQL.
type ExaminationWriteRepository struct{ store *Store }

// NewExaminationWriteRepository создаёт репозиторий записи обследований.
func NewExaminationWriteRepository(store *Store) *ExaminationWriteRepository {
	return &ExaminationWriteRepository{store: store}
}

func (r *ExaminationWriteRepository) transaction(ctx context.Context, operation func(pgx.Tx) error) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := operation(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertOutbox(ctx context.Context, tx pgx.Tx, aggregateID uuid.UUID, eventType string, payload []byte, createdAt time.Time) error {
	_, err := tx.Exec(ctx, sqlqueries.OutboxCreate, uuid.New(), aggregateID, eventType, payload, createdAt)
	return err
}

func (r *ExaminationWriteRepository) CreateExamination(ctx context.Context, examination ports.ExaminationWriteModel, job ports.ProcessingJobWriteModel) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.PatientCreate, examination.PatientID, examination.CreatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationCreate, examination.ID, examination.DoctorID, examination.PatientID, examination.ExaminationDate, examination.Status, examination.CreatedAt, examination.UpdatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingJobCreate, job.ID, job.ExaminationID, job.Status, job.Attempt, job.CreatedAt, job.UpdatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examination.ID, "examination.created", []byte(`{"status":"created"}`), examination.CreatedAt)
	})
}

func (r *ExaminationWriteRepository) StartProcessing(ctx context.Context, examinationID, jobID uuid.UUID, startedAt time.Time, attempt int) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationStart, examinationID, startedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingJobStart, jobID, startedAt, attempt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.processing", []byte(`{"status":"processing"}`), startedAt)
	})
}

func (r *ExaminationWriteRepository) SaveTranscript(ctx context.Context, examinationID, jobID uuid.UUID, transcript string, updatedAt time.Time) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.TranscriptUpsert, examinationID, transcript, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationMarkTranscribed, examinationID, updatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.transcribed", []byte(`{"status":"transcribed"}`), updatedAt)
	})
}

func (r *ExaminationWriteRepository) SaveSummary(ctx context.Context, examinationID uuid.UUID, summary string, updatedAt time.Time) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.SummaryUpsert, examinationID, summary, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationMarkSummarized, examinationID, updatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.summarized", []byte(`{"status":"summarized"}`), updatedAt)
	})
}

func (r *ExaminationWriteRepository) CompleteProcessing(ctx context.Context, examinationID, jobID uuid.UUID, updatedAt time.Time) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationComplete, examinationID, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingJobComplete, jobID, updatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.completed", []byte(`{"status":"completed"}`), updatedAt)
	})
}

func (r *ExaminationWriteRepository) FailProcessing(ctx context.Context, examinationID, jobID uuid.UUID, errText string, updatedAt time.Time) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationFail, examinationID, errText, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingJobFail, jobID, errText, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingErrorCreate, uuid.New(), examinationID, jobID, errText, updatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.failed", []byte(`{"status":"failed"}`), updatedAt)
	})
}

func (r *ExaminationWriteRepository) RetryProcessing(ctx context.Context, doctorID int64, examinationID, _ uuid.UUID, updatedAt time.Time) (uuid.UUID, string, error) {
	var jobID uuid.UUID
	var transcript string
	err := r.transaction(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, sqlqueries.RetryJobGet, examinationID, doctorID).Scan(&jobID, &transcript); err != nil {
			return err
		}
		if transcript == "" {
			return ports.ErrFileRequired
		}
		if _, err := tx.Exec(ctx, sqlqueries.ProcessingJobReset, jobID, updatedAt); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sqlqueries.ExaminationReset, examinationID, updatedAt); err != nil {
			return err
		}
		return insertOutbox(ctx, tx, examinationID, "examination.retry", []byte(`{"status":"created"}`), updatedAt)
	})
	return jobID, transcript, err
}

func (r *ExaminationWriteRepository) DeleteExamination(ctx context.Context, doctorID int64, examinationID uuid.UUID) error {
	return r.transaction(ctx, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, sqlqueries.ExaminationDelete, examinationID, doctorID)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return ports.ErrExaminationNotFound
		}
		payload := []byte(fmt.Sprintf(`{"doctor_id":%d}`, doctorID))
		return insertOutbox(ctx, tx, examinationID, "examination.deleted", payload, time.Now().UTC())
	})
}

var _ ports.ExaminationWriteRepository = (*ExaminationWriteRepository)(nil)

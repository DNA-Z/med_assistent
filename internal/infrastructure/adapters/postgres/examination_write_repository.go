package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type ExaminationWriteRepository struct {
	store *Store
}

func NewExaminationWriteRepository(
	store *Store,
) *ExaminationWriteRepository {
	return &ExaminationWriteRepository{
		store: store,
	}
}

func (r *ExaminationWriteRepository) CreateExamination(
	ctx context.Context,
	examination ports.ExaminationWriteModel,
	job ports.ProcessingJobWriteModel,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO examinations (
			id,
			doctor_id,
			patient_id,
			examination_date,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		examination.ID,
		examination.DoctorID,
		examination.PatientID,
		examination.ExaminationDate,
		examination.Status,
		examination.CreatedAt,
		examination.UpdatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO processing_jobs (
			id,
			examination_id,
			status,
			attempt,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		job.ID,
		job.ExaminationID,
		job.Status,
		job.Attempt,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examination.ID,
		"examination.created",
		[]byte(`{"status":"created"}`),
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) StartProcessing(
	ctx context.Context,
	examinationID uuid.UUID,
	jobID uuid.UUID,
	startedAt time.Time,
	attempt int,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'processing',
		     updated_at = $2,
		     error_reason = NULL
		 WHERE id = $1`,
		examinationID,
		startedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE processing_jobs
		 SET status = 'processing',
		     attempt = $3,
		     started_at = $2,
		     completed_at = NULL,
		     error_reason = NULL,
		     updated_at = $2
		 WHERE id = $1`,
		jobID,
		startedAt,
		attempt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) SaveTranscript(
	ctx context.Context,
	examinationID uuid.UUID,
	jobID uuid.UUID,
	transcript string,
	updatedAt time.Time,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO transcripts (
			examination_id,
			text,
			created_at
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (examination_id)
		DO UPDATE SET text = EXCLUDED.text`,
		examinationID,
		transcript,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'transcribed',
		     updated_at = $2
		 WHERE id = $1`,
		examinationID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		"examination.transcribed",
		[]byte(`{"status":"transcribed"}`),
		updatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) SaveSummary(
	ctx context.Context,
	examinationID uuid.UUID,
	summary string,
	updatedAt time.Time,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO summaries (
			examination_id,
			text,
			created_at
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (examination_id)
		DO UPDATE SET text = EXCLUDED.text`,
		examinationID,
		summary,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'summarized',
		     updated_at = $2
		 WHERE id = $1`,
		examinationID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		"examination.summarized",
		[]byte(`{"status":"summarized"}`),
		updatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) CompleteProcessing(
	ctx context.Context,
	examinationID uuid.UUID,
	jobID uuid.UUID,
	updatedAt time.Time,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'completed',
		     updated_at = $2
		 WHERE id = $1`,
		examinationID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE processing_jobs
		 SET status = 'completed',
		     completed_at = $2,
		     updated_at = $2
		 WHERE id = $1`,
		jobID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		"examination.completed",
		[]byte(`{"status":"completed"}`),
		updatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) FailProcessing(
	ctx context.Context,
	examinationID uuid.UUID,
	jobID uuid.UUID,
	errText string,
	updatedAt time.Time,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'failed',
		     error_reason = $2,
		     updated_at = $3
		 WHERE id = $1`,
		examinationID,
		errText,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE processing_jobs
		 SET status = 'failed',
		     error_reason = $2,
		     updated_at = $3
		 WHERE id = $1`,
		jobID,
		errText,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO processing_errors (
			id,
			examination_id,
			processing_job_id,
			error,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		jobID,
		errText,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		"examination.failed",
		[]byte(`{"status":"failed"}`),
		updatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) RetryProcessing(
	ctx context.Context,
	examinationID uuid.UUID,
	jobID uuid.UUID,
	updatedAt time.Time,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldJobID uuid.UUID

	err = tx.QueryRow(
		ctx,
		`SELECT id
		 FROM processing_jobs
		 WHERE examination_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		examinationID,
	).Scan(&oldJobID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE processing_jobs
		 SET status = 'created',
		     error_reason = NULL,
		     completed_at = NULL,
		     updated_at = $2
		 WHERE id = $1
		   AND status = 'failed'`,
		oldJobID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE examinations
		 SET status = 'created',
		     error_reason = NULL,
		     updated_at = $2
		 WHERE id = $1
		   AND status = 'failed'`,
		examinationID,
		updatedAt,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			id,
			aggregate_id,
			event_type,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(),
		examinationID,
		"examination.retry",
		[]byte(`{"status":"created"}`),
		updatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ExaminationWriteRepository) DeleteExamination(
	ctx context.Context,
	doctorID int64,
	examinationID uuid.UUID,
) error {
	tx, err := r.store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(
		ctx,
		`DELETE FROM examinations
		 WHERE id = $1
		   AND doctor_id = $2`,
		examinationID,
		doctorID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	return tx.Commit(ctx)
}

var _ ports.ExaminationWriteRepository = (*ExaminationWriteRepository)(nil)

package redis

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type OutboxWorker struct {
	pg     *pgxpool.Pool
	redis  *Client
	logger *slog.Logger
}

func NewOutboxWorker(
	pg *pgxpool.Pool,
	redisClient *Client,
	logger *slog.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		pg:     pg,
		redis:  redisClient,
		logger: logger,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := w.process(ctx); err != nil {
				w.logger.Error(
					"failed to process outbox",
					"error", err,
				)
			}
		}
	}
}

func (w *OutboxWorker) process(
	ctx context.Context,
) error {
	tx, err := w.pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(
		ctx,
		`SELECT
			id,
			aggregate_id,
			event_type,
			payload
		FROM outbox_events
		WHERE processed_at IS NULL
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 100`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			eventID     uuid.UUID
			aggregateID uuid.UUID
			eventType   string
			payload     []byte
		)

		if err := rows.Scan(
			&eventID,
			&aggregateID,
			&eventType,
			&payload,
		); err != nil {
			return err
		}

		if err := w.project(
			ctx,
			aggregateID,
		); err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`UPDATE outbox_events
			 SET processed_at = now()
			 WHERE id = $1`,
			eventID,
		)
		if err != nil {
			return err
		}

		_ = eventType
		_ = payload
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (w *OutboxWorker) project(
	ctx context.Context,
	examinationID uuid.UUID,
) error {
	var item ports.ExaminationDTO

	err := w.pg.QueryRow(
		ctx,
		`SELECT
			e.id,
			e.doctor_id,
			e.patient_id,
			e.examination_date,
			e.status,
			COALESCE(t.text, ''),
			COALESCE(s.text, ''),
			COALESCE(d.text, ''),
			COALESCE(e.error_reason, ''),
			e.created_at,
			e.updated_at
		FROM examinations e
		LEFT JOIN transcripts t
			ON t.examination_id = e.id
		LEFT JOIN summaries s
			ON s.examination_id = e.id
		LEFT JOIN diagnoses d
			ON d.examination_id = e.id
		WHERE e.id = $1`,
		examinationID,
	).Scan(
		&item.ID,
		&item.DoctorID,
		&item.PatientID,
		&item.ExaminationDate,
		&item.Status,
		&item.Transcript,
		&item.Summary,
		&item.Diagnosis,
		&item.ErrorReason,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return StoreExamination(
		ctx,
		w.redis,
		item,
	)
}

func DecodePayload(data []byte) map[string]any {
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

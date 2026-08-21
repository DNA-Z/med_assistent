package redis

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	postgresadapter "github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres"
)

type OutboxWorker struct {
	pg     *pgxpool.Pool
	redis  *Client
	logger *slog.Logger
}

type outboxEvent struct {
	id          uuid.UUID
	aggregateID uuid.UUID
	eventType   string
	payload     []byte
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

	repository := postgresadapter.NewQueryRepository(
		tx,
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
		func(row pgx.CollectableRow) (outboxEvent, error) {
			var event outboxEvent
			err := row.Scan(&event.id, &event.aggregateID, &event.eventType, &event.payload)
			return event, err
		},
	)

	for result := range repository.Seq(ctx) {
		if result.Err != nil {
			return result.Err
		}
		event := result.Value
		if event.eventType == "examination.deleted" {
			var deleted struct {
				DoctorID int64 `json:"doctor_id"`
			}
			if err := json.Unmarshal(event.payload, &deleted); err != nil {
				return err
			}
			if err := RemoveExamination(ctx, w.redis, deleted.DoctorID, event.aggregateID); err != nil {
				return err
			}
		} else if err := w.project(ctx, event.aggregateID); err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			`UPDATE outbox_events
			 SET processed_at = now()
			 WHERE id = $1`,
			event.id,
		)
		if err != nil {
			return err
		}
		w.logger.Debug("outbox-событие обработано", "event_id", event.id, "event_type", event.eventType, "aggregate_id", event.aggregateID)

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

	return ProjectExamination(
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

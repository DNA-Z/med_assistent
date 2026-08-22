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
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres/sqlqueries"
)

type OutboxWorker struct {
	pg      *pgxpool.Pool
	redis   *Client
	logger  *slog.Logger
	storage ports.ObjectStorage
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
	storage ports.ObjectStorage,
) *OutboxWorker {
	return &OutboxWorker{
		pg:      pg,
		redis:   redisClient,
		logger:  logger,
		storage: storage,
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
		sqlqueries.OutboxBatch,
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
				DoctorID  int64  `json:"doctor_id"`
				ObjectKey string `json:"object_key"`
			}
			if err := json.Unmarshal(event.payload, &deleted); err != nil {
				return err
			}
			if err := RemoveExamination(ctx, w.redis, deleted.DoctorID, event.aggregateID); err != nil {
				return err
			}
			if err := w.storage.Delete(ctx, deleted.ObjectKey); err != nil {
				return err
			}
		} else if err := w.project(ctx, event.aggregateID); err != nil {
			return err
		}

		_, err = tx.Exec(
			ctx,
			sqlqueries.OutboxMarkProcessed,
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
		sqlqueries.ExaminationProjectionGet,
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

package event_handlers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type OutboxEventHandler interface {
	Handle(ctx context.Context, event OutboxEvent) error
}

type OutboxEvent struct {
	ID          uuid.UUID
	AggregateID uuid.UUID
	Type        string
	Payload     []byte
}

type Logger struct {
	logger *slog.Logger
}

func NewLogger(logger *slog.Logger) *Logger {
	return &Logger{logger: logger}
}

func (h *Logger) Handle(
	_ context.Context,
	event OutboxEvent,
) error {
	h.logger.Debug(
		"outbox event",
		"event_id", event.ID,
		"aggregate_id", event.AggregateID,
		"type", event.Type,
	)

	return nil
}

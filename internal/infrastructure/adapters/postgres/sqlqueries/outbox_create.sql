INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at)
VALUES ($1, $2, $3, $4, $5);

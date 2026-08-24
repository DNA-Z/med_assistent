UPDATE outbox_events
SET processed_at = now()
WHERE id = $1;

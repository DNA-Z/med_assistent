SELECT id, aggregate_id, event_type, payload
FROM outbox_events
WHERE processed_at IS NULL
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 100;

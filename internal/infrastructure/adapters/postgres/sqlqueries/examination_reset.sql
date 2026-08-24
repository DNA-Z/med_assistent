UPDATE examinations
SET status = 'created', error_reason = NULL, updated_at = $2
WHERE id = $1 AND status = 'failed';

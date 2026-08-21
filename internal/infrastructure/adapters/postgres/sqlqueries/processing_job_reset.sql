UPDATE processing_jobs
SET status = 'created', error_reason = NULL, completed_at = NULL, updated_at = $2
WHERE id = $1 AND status = 'failed';

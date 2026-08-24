UPDATE processing_jobs
SET status = 'processing', attempt = $3, started_at = $2,
    completed_at = NULL, error_reason = NULL, updated_at = $2
WHERE id = $1;

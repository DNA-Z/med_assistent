UPDATE processing_jobs
SET status = 'failed', error_reason = $2, updated_at = $3
WHERE id = $1;

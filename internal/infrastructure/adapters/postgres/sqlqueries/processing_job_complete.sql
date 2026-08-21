UPDATE processing_jobs
SET status = 'completed', completed_at = $2, updated_at = $2
WHERE id = $1;

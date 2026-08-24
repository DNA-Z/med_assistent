UPDATE examinations
SET status = 'processing', updated_at = $2, error_reason = NULL
WHERE id = $1;

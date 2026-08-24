INSERT INTO processing_errors (id, examination_id, processing_job_id, error, created_at)
VALUES ($1, $2, $3, $4, $5);

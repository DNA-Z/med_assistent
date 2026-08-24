SELECT
    j.id,
    j.examination_id,
    j.attempt,
    COALESCE(t.text, ''),
    COALESCE(e.audio_object_key, ''),
    COALESCE(e.audio_file_name, '')
FROM processing_jobs j
JOIN examinations e ON e.id = j.examination_id
LEFT JOIN transcripts t ON t.examination_id = e.id
WHERE j.status = 'created'
   OR (
       j.status = 'processing'
       AND (j.started_at IS NULL OR j.started_at < $1)
   )
ORDER BY j.created_at
FOR UPDATE OF j SKIP LOCKED
LIMIT $2;

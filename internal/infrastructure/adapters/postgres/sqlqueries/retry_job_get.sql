SELECT j.id, COALESCE(t.text, '')
FROM processing_jobs j
JOIN examinations e ON e.id = j.examination_id
LEFT JOIN transcripts t ON t.examination_id = e.id
WHERE j.examination_id = $1 AND e.doctor_id = $2 AND e.status = 'failed'
ORDER BY j.created_at DESC
LIMIT 1;

SELECT e.id, e.doctor_id, e.patient_id, e.examination_date, e.status,
       COALESCE(t.text, ''), COALESCE(s.text, ''), COALESCE(d.text, ''),
       COALESCE(e.error_reason, ''), e.created_at, e.updated_at
FROM examinations e
LEFT JOIN transcripts t ON t.examination_id = e.id
LEFT JOIN summaries s ON s.examination_id = e.id
LEFT JOIN diagnoses d ON d.examination_id = e.id
WHERE e.id = $1;

INSERT INTO summaries (examination_id, text, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (examination_id) DO UPDATE SET text = EXCLUDED.text;

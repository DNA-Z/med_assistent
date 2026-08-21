INSERT INTO doctors (telegram_id, created_at)
VALUES ($1, now())
ON CONFLICT (telegram_id) DO NOTHING;

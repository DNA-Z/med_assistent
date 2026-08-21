INSERT INTO patients (id, created_at)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;

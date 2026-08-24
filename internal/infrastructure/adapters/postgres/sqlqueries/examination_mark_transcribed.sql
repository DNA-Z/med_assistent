UPDATE examinations SET status = 'transcribed', updated_at = $2 WHERE id = $1;

UPDATE examinations SET status = 'completed', updated_at = $2 WHERE id = $1;

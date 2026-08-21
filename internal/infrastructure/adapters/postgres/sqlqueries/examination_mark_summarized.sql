UPDATE examinations SET status = 'summarized', updated_at = $2 WHERE id = $1;

SELECT EXISTS (SELECT 1 FROM doctors WHERE telegram_id = $1);

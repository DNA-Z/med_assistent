DELETE FROM examinations WHERE id = $1 AND doctor_id = $2
RETURNING COALESCE(audio_object_key, '');

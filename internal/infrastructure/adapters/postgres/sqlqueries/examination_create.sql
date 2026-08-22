INSERT INTO examinations (
    id, doctor_id, patient_id, examination_date, status, created_at, updated_at,
    audio_object_key, audio_file_name, audio_content_type, audio_size
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

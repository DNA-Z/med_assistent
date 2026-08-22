ALTER TABLE examinations
    ADD COLUMN audio_object_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN audio_file_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN audio_content_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN audio_size BIGINT NOT NULL DEFAULT 0 CHECK (audio_size >= 0);

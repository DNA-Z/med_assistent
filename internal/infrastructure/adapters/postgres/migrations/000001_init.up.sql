CREATE TABLE doctors (
    telegram_id BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE patients (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE examinations (
    id UUID PRIMARY KEY,
    doctor_id BIGINT NOT NULL REFERENCES doctors(telegram_id),
    patient_id UUID NOT NULL REFERENCES patients(id),
    examination_date TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('created','processing','transcribed','summarized','completed','failed')),
    error_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX examinations_doctor_created_idx ON examinations (doctor_id, created_at DESC);
CREATE INDEX examinations_status_idx ON examinations (status);

CREATE TABLE processing_jobs (
    id UUID PRIMARY KEY,
    examination_id UUID NOT NULL REFERENCES examinations(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    attempt INTEGER NOT NULL DEFAULT 0,
    error_reason TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX processing_jobs_examination_idx ON processing_jobs (examination_id, created_at DESC);
CREATE INDEX processing_jobs_status_idx ON processing_jobs (status);

CREATE TABLE transcripts (
    examination_id UUID PRIMARY KEY REFERENCES examinations(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX transcripts_search_idx ON transcripts USING GIN (to_tsvector('simple', text));

CREATE TABLE summaries (
    examination_id UUID PRIMARY KEY REFERENCES examinations(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE diagnoses (
    examination_id UUID PRIMARY KEY REFERENCES examinations(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE processing_errors (
    id UUID PRIMARY KEY,
    examination_id UUID NOT NULL REFERENCES examinations(id) ON DELETE CASCADE,
    processing_job_id UUID NOT NULL REFERENCES processing_jobs(id) ON DELETE CASCADE,
    error TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE INDEX outbox_unprocessed_idx ON outbox_events (created_at) WHERE processed_at IS NULL;

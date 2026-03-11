CREATE TABLE videos (
    -- id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    description      TEXT,
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    raw_blob_url     TEXT,
    duration_seconds INTEGER,
    file_size_bytes  BIGINT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

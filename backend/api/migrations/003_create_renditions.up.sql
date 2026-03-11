CREATE TABLE hls_renditions (
    -- id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id     UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    resolution   TEXT NOT NULL CHECK (resolution IN ('360p', '480p', '720p', '1080p')),
    playlist_url TEXT NOT NULL,
    bandwidth    INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    -- id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

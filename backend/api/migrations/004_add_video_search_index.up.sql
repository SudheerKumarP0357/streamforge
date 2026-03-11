CREATE INDEX idx_videos_title ON videos USING gin(to_tsvector('english', title));

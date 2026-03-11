package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"streamforge/api/internal/models"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Ping(ctx context.Context) error {
	var result int
	return db.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// User methods

func (db *DB) CreateUser(ctx context.Context, email, passwordHash string) (string, error) {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	var id string
	err := db.pool.QueryRow(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
	var user models.User
	err := db.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (db *DB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	var user models.User
	err := db.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// Video methods

func (db *DB) CreateVideo(ctx context.Context, userID, title, description string) (string, error) {
	query := `INSERT INTO videos (user_id, title, description, status) VALUES ($1, $2, $3, 'pending') RETURNING id`
	var id string
	
	var descPtr *string
	if description != "" {
		descPtr = &description
	}
	
	err := db.pool.QueryRow(ctx, query, userID, title, descPtr).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create video: %w", err)
	}
	return id, nil
}

func (db *DB) GetVideoByID(ctx context.Context, id string) (*models.Video, error) {
	query := `SELECT id, user_id, title, description, status, raw_blob_url, duration_seconds, file_size_bytes, created_at, updated_at FROM videos WHERE id = $1`
	var v models.Video
	err := db.pool.QueryRow(ctx, query, id).Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.Status, &v.RawBlobURL, &v.DurationSeconds, &v.FileSizeBytes, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get video by id: %w", err)
	}
	
	// Preload renditions mapped by video
	renditions, _ := db.GetRenditionsByVideoID(ctx, id)
	v.Renditions = renditions

	return &v, nil
}

func (db *DB) ListVideosByUser(ctx context.Context, userID string, page, limit int) ([]models.Video, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	countQuery := `SELECT count(*) FROM videos WHERE user_id = $1`
	err := db.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count videos: %w", err)
	}

	query := `SELECT id, user_id, title, description, status, raw_blob_url, duration_seconds, file_size_bytes, created_at, updated_at 
			  FROM videos WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := db.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list videos: %w", err)
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var v models.Video
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.Status, &v.RawBlobURL, &v.DurationSeconds, &v.FileSizeBytes, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}
		
		// Load renditions to match expected full structure
		renditions, _ := db.GetRenditionsByVideoID(ctx, v.ID)
		v.Renditions = renditions
		
		videos = append(videos, v)
	}

	return videos, total, nil
}

func (db *DB) SearchVideosByUser(ctx context.Context, userID, searchQuery string, page, limit int) ([]models.Video, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	countQuery := `SELECT count(*) FROM videos 
				   WHERE user_id = $1 
				   AND (to_tsvector('english', title) @@ plainto_tsquery('english', $2) OR title ILIKE '%' || $2 || '%')`
	err := db.pool.QueryRow(ctx, countQuery, userID, searchQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search videos: %w", err)
	}

	query := `SELECT id, user_id, title, description, status, raw_blob_url, duration_seconds, file_size_bytes, created_at, updated_at 
			  FROM videos 
			  WHERE user_id = $1 
			  AND (to_tsvector('english', title) @@ plainto_tsquery('english', $2) OR title ILIKE '%' || $2 || '%')
			  ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	rows, err := db.pool.Query(ctx, query, userID, searchQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search videos: %w", err)
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var v models.Video
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.Status, &v.RawBlobURL, &v.DurationSeconds, &v.FileSizeBytes, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search video: %w", err)
		}
		
		// Load renditions to match expected full structure
		renditions, _ := db.GetRenditionsByVideoID(ctx, v.ID)
		v.Renditions = renditions
		
		videos = append(videos, v)
	}

	return videos, total, nil
}

func (db *DB) ListAllVideosAdmin(ctx context.Context, page, limit int) ([]models.AdminVideo, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM videos`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count all videos: %w", err)
	}

	query := `SELECT v.id, v.user_id, v.title, v.description, v.status, v.raw_blob_url, v.duration_seconds, v.file_size_bytes, v.created_at, v.updated_at, u.email 
			  FROM videos v 
			  JOIN users u ON v.user_id = u.id 
			  ORDER BY v.created_at DESC 
			  LIMIT $1 OFFSET $2`
	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all admin videos: %w", err)
	}
	defer rows.Close()

	var videos []models.AdminVideo
	for rows.Next() {
		var av models.AdminVideo
		err := rows.Scan(&av.ID, &av.UserID, &av.Title, &av.Description, &av.Status, &av.RawBlobURL, &av.DurationSeconds, &av.FileSizeBytes, &av.CreatedAt, &av.UpdatedAt, &av.UserEmail)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan admin video: %w", err)
		}
		
		// Load renditions to match expected full structure
		renditions, _ := db.GetRenditionsByVideoID(ctx, av.ID)
		av.Renditions = renditions
		
		videos = append(videos, av)
	}

	return videos, total, nil
}

func (db *DB) GetAdminStats(ctx context.Context) (*models.AdminStats, error) {
	var stats models.AdminStats

	// Fast concurrent queries or simple serial queries since it's an admin panel
	// We'll just do serial for simplicity right now
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("admin stats - failed to count users: %w", err)
	}

	err = db.pool.QueryRow(ctx, `SELECT count(*) FROM videos`).Scan(&stats.TotalVideos)
	if err != nil {
		return nil, fmt.Errorf("admin stats - failed to count videos: %w", err)
	}

	// Group by status
	rows, err := db.pool.Query(ctx, `SELECT status, count(*) FROM videos GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("admin stats - failed to group video stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("admin stats - failed to scan status count: %w", err)
		}

		switch status {
		case string(models.StatusPending):
			stats.VideosByStatus.Pending = count
		case string(models.StatusProcessing):
			stats.VideosByStatus.Processing = count
		case string(models.StatusReady):
			stats.VideosByStatus.Ready = count
		case string(models.StatusFailed):
			stats.VideosByStatus.Failed = count
		}
	}

	return &stats, nil
}

func (db *DB) UpdateVideoStatus(ctx context.Context, videoID, status string) error {
	query := `UPDATE videos SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.pool.Exec(ctx, query, status, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}
	return nil
}

func (db *DB) UpdateVideoRawBlobURL(ctx context.Context, videoID, blobURL string) error {
	query := `UPDATE videos SET raw_blob_url = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.pool.Exec(ctx, query, blobURL, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video raw blob url: %w", err)
	}
	return nil
}

func (db *DB) DeleteVideo(ctx context.Context, id string) error {
	query := `DELETE FROM videos WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}
	return nil
}

// Rendition methods

func (db *DB) CreateRendition(ctx context.Context, videoID, resolution, playlistURL string, bandwidth int) error {
	query := `INSERT INTO hls_renditions (video_id, resolution, playlist_url, bandwidth) VALUES ($1, $2, $3, $4)`
	_, err := db.pool.Exec(ctx, query, videoID, resolution, playlistURL, bandwidth)
	if err != nil {
		return fmt.Errorf("failed to create rendition: %w", err)
	}
	return nil
}

func (db *DB) GetRenditionsByVideoID(ctx context.Context, videoID string) ([]models.HLSRendition, error) {
	query := `SELECT id, video_id, resolution, playlist_url, bandwidth, created_at FROM hls_renditions WHERE video_id = $1 ORDER BY bandwidth DESC`
	rows, err := db.pool.Query(ctx, query, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get renditions: %w", err)
	}
	defer rows.Close()

	var renditions []models.HLSRendition
	for rows.Next() {
		var r models.HLSRendition
		err := rows.Scan(&r.ID, &r.VideoID, &r.Resolution, &r.PlaylistURL, &r.Bandwidth, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rendition: %w", err)
		}
		renditions = append(renditions, r)
	}

	return renditions, nil
}

// Refresh token methods

func (db *DB) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := db.pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (db *DB) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, revoked, created_at FROM refresh_tokens WHERE token_hash = $1`
	var token models.RefreshToken
	err := db.pool.QueryRow(ctx, query, tokenHash).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.Revoked, &token.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &token, nil
}

func (db *DB) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`
	_, err := db.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

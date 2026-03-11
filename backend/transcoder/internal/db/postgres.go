package db

import (
	"context"
	"fmt"
	"log"

	"streamforge/transcoder/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	pool *pgxpool.Pool
}

type HLSRendition struct {
	VideoID     string
	Resolution  string
	PlaylistURL string
	Bandwidth   int
}

func NewPostgresClient(cfg config.Config) *PostgresClient {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB, cfg.PostgresSSLMode)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return &PostgresClient{pool: pool}
}

// UpdateVideoStatus updates the status of a video in the database
func (db *PostgresClient) UpdateVideoStatus(ctx context.Context, videoID string, status string) error {
	query := `UPDATE videos SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.pool.Exec(ctx, query, status, videoID)
	return err
}

// InsertRenditions inserts the generated rendition metadata
func (db *PostgresClient) InsertRenditions(ctx context.Context, renditions []HLSRendition) error {
	if len(renditions) == 0 {
		return nil
	}

	query := `INSERT INTO hls_renditions (video_id, resolution, playlist_url, bandwidth) VALUES `
	args := []interface{}{}
	
	for i, r := range renditions {
		if i > 0 {
			query += ", "
		}
		// Calculate precise param indexes based on iteration
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		args = append(args, r.VideoID, r.Resolution, r.PlaylistURL, r.Bandwidth)
	}

	_, err := db.pool.Exec(ctx, query, args...)
	return err
}

func (db *PostgresClient) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

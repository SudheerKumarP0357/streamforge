package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"streamforge/transcoder/internal/consumer"
	"streamforge/transcoder/internal/db"
	"streamforge/transcoder/internal/ffmpeg"
	"streamforge/transcoder/internal/metrics"
	"streamforge/transcoder/internal/storage"
)

type Pipeline struct {
	consumer   *consumer.Consumer
	transcoder *ffmpeg.Transcoder
	storage    *storage.BlobClient
	db         *db.PostgresClient
}

func NewPipeline(c *consumer.Consumer, t *ffmpeg.Transcoder, s *storage.BlobClient, d *db.PostgresClient) *Pipeline {
	return &Pipeline{
		consumer:   c,
		transcoder: t,
		storage:    s,
		db:         d,
	}
}

func (p *Pipeline) Process(ctx context.Context, job consumer.TranscodeJob) error {
	start := time.Now()
	logger := slog.With("video_id", job.VideoID)
	logger.Info("Starting pipeline")

	// Ensure cleanup
	defer func() {
		logger.Info("Cleaning up temporary files")
		if err := p.transcoder.Cleanup(job.VideoID); err != nil {
			logger.Error("Failed to cleanup temp dir", "error", err)
		}
	}()

	// Temporary directory for the video
	tempDir := filepath.Join(p.transcoder.TempDir, job.VideoID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to create temp dir: %w", err))
	}

	// Step 1: Download raw file
	inputPath := filepath.Join(tempDir, "input.mp4")
	blobName := fmt.Sprintf("%s.mp4", job.VideoID)
	logger.Info("[Step 1] Downloading raw video from blob", "blob_name", blobName)
	if err := p.storage.DownloadBlob(ctx, blobName, inputPath); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to download raw video: %w", err))
	}

	// Step 2: Update status to 'processing'
	logger.Info("[Step 2] Updating status to processing")
	if err := p.db.UpdateVideoStatus(ctx, job.VideoID, "processing"); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to update status to processing: %w", err))
	}

	// Step 3: Transcode
	logger.Info("[Step 3] Transcoding video")
	result, err := p.transcoder.Transcode(ctx, job.VideoID, inputPath)
	if err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("transcoding failed: %w", err))
	}

	// Step 4: Generate master playlist
	logger.Info("[Step 4] Generating master playlist")
	if err := p.transcoder.GenerateMasterPlaylist(result, result.OutputDir); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to generate master playlist: %w", err))
	}

	// Step 5: Upload HLS files
	logger.Info("[Step 5] Uploading HLS files")
	// Master playlist first
	masterPath := filepath.Join(result.OutputDir, "master.m3u8")
	_, err = p.storage.UploadBlob(ctx, job.VideoID, masterPath, "application/x-mpegURL")
	if err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to upload master playlist: %w", err))
	}

	renditionsDB := make([]db.HLSRendition, 0, len(result.Renditions))

	for _, rend := range result.Renditions {
		// Upload rendition playlist (.m3u8)
		playlistURL, err := p.storage.UploadBlob(ctx, job.VideoID, rend.PlaylistFile, "application/x-mpegURL")
		if err != nil {
			return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to upload playlist %s: %w", rend.PlaylistFile, err))
		}

		// Upload all segments (.ts)
		for _, seg := range rend.SegmentFiles {
			_, err := p.storage.UploadBlob(ctx, job.VideoID, seg, "video/MP2T")
			if err != nil {
				return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to upload segment %s: %w", seg, err))
			}
		}

		renditionsDB = append(renditionsDB, db.HLSRendition{
			VideoID:     job.VideoID,
			Resolution:  rend.Resolution,
			PlaylistURL: playlistURL,
			Bandwidth:   rend.Bandwidth,
		})
	}

	// Step 6: Insert HLS renditions into DB
	logger.Info("[Step 6] Inserting HLS renditions")
	if err := p.db.InsertRenditions(ctx, renditionsDB); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to insert renditions: %w", err))
	}

	// Step 7: Update status to 'ready'
	logger.Info("[Step 7] Updating status to ready")
	if err := p.db.UpdateVideoStatus(ctx, job.VideoID, "ready"); err != nil {
		return p.fail(ctx, logger, job.VideoID, fmt.Errorf("failed to update status to ready: %w", err))
	}

	logger.Info("Pipeline completed successfully")
	metrics.TranscodeDuration.Observe(time.Since(start).Seconds())
	metrics.TranscodeJobsTotal.WithLabelValues("success").Inc()
	return nil
}

func (p *Pipeline) fail(ctx context.Context, logger *slog.Logger, videoID string, err error) error {
	logger.Error("Pipeline failed", "error", err)
	// Attempt to mark as failed in DB
	if updateErr := p.db.UpdateVideoStatus(ctx, videoID, "failed"); updateErr != nil {
		logger.Error("Critical: also failed to update video status to failed", "update_err", updateErr)
	}
	metrics.TranscodeJobsTotal.WithLabelValues("failed").Inc()
	return err
}

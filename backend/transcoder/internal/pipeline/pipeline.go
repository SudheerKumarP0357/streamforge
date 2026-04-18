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
	pipelineStart := time.Now()
	logger := slog.With("component", "pipeline", "video_id", job.VideoID)
	logger.Info("pipeline started", "job_received_at", job.UploadedAt)

	// Ensure cleanup
	defer func() {
		logger.Info("cleaning up temporary files")
		if err := p.transcoder.Cleanup(job.VideoID); err != nil {
			logger.Error("failed to cleanup temp dir", "error", err.Error())
		}
	}()

	// Temporary directory for the video
	tempDir := filepath.Join(p.transcoder.TempDir, job.VideoID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return p.fail(ctx, logger, job.VideoID, "create_temp_dir", pipelineStart, fmt.Errorf("failed to create temp dir: %w", err))
	}

	// Step 1: Download raw file
	stepStart := time.Now()
	inputPath := filepath.Join(tempDir, "input.mp4")
	blobName := fmt.Sprintf("%s.mp4", job.VideoID)
	logger.Info("step starting", "step_number", 1, "step_name", "download_raw_video", "blob_name", blobName)
	if err := p.storage.DownloadBlob(ctx, blobName, inputPath); err != nil {
		return p.fail(ctx, logger, job.VideoID, "download_raw_video", pipelineStart, fmt.Errorf("failed to download raw video: %w", err))
	}
	logger.Info("step completed", "step_number", 1, "step_name", "download_raw_video", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 2: Update status to 'processing'
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 2, "step_name", "update_status_processing")
	if err := p.db.UpdateVideoStatus(ctx, job.VideoID, "processing"); err != nil {
		return p.fail(ctx, logger, job.VideoID, "update_status_processing", pipelineStart, fmt.Errorf("failed to update status to processing: %w", err))
	}
	logger.Info("step completed", "step_number", 2, "step_name", "update_status_processing", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 3: Transcode
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 3, "step_name", "transcode")
	result, err := p.transcoder.Transcode(ctx, job.VideoID, inputPath)
	if err != nil {
		return p.fail(ctx, logger, job.VideoID, "transcode", pipelineStart, fmt.Errorf("transcoding failed: %w", err))
	}
	logger.Info("step completed", "step_number", 3, "step_name", "transcode", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 4: Generate master playlist
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 4, "step_name", "generate_master_playlist")
	if err := p.transcoder.GenerateMasterPlaylist(result, result.OutputDir); err != nil {
		return p.fail(ctx, logger, job.VideoID, "generate_master_playlist", pipelineStart, fmt.Errorf("failed to generate master playlist: %w", err))
	}
	logger.Info("step completed", "step_number", 4, "step_name", "generate_master_playlist", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 5: Upload HLS files
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 5, "step_name", "upload_hls_files")
	// Master playlist first
	masterPath := filepath.Join(result.OutputDir, "master.m3u8")
	_, err = p.storage.UploadBlob(ctx, job.VideoID, masterPath, "application/x-mpegURL")
	if err != nil {
		return p.fail(ctx, logger, job.VideoID, "upload_hls_files", pipelineStart, fmt.Errorf("failed to upload master playlist: %w", err))
	}

	renditionsDB := make([]db.HLSRendition, 0, len(result.Renditions))

	for _, rend := range result.Renditions {
		// Upload rendition playlist (.m3u8)
		playlistURL, err := p.storage.UploadBlob(ctx, job.VideoID, rend.PlaylistFile, "application/x-mpegURL")
		if err != nil {
			return p.fail(ctx, logger, job.VideoID, "upload_hls_files", pipelineStart, fmt.Errorf("failed to upload playlist %s: %w", rend.PlaylistFile, err))
		}

		// Upload all segments (.ts)
		for _, seg := range rend.SegmentFiles {
			_, err := p.storage.UploadBlob(ctx, job.VideoID, seg, "video/MP2T")
			if err != nil {
				return p.fail(ctx, logger, job.VideoID, "upload_hls_files", pipelineStart, fmt.Errorf("failed to upload segment %s: %w", seg, err))
			}
		}

		renditionsDB = append(renditionsDB, db.HLSRendition{
			VideoID:     job.VideoID,
			Resolution:  rend.Resolution,
			PlaylistURL: playlistURL,
			Bandwidth:   rend.Bandwidth,
		})
	}
	logger.Info("step completed", "step_number", 5, "step_name", "upload_hls_files", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 6: Insert HLS renditions into DB
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 6, "step_name", "insert_renditions")
	if err := p.db.InsertRenditions(ctx, renditionsDB); err != nil {
		return p.fail(ctx, logger, job.VideoID, "insert_renditions", pipelineStart, fmt.Errorf("failed to insert renditions: %w", err))
	}
	logger.Info("step completed", "step_number", 6, "step_name", "insert_renditions", "duration_ms", time.Since(stepStart).Milliseconds())

	// Step 7: Update status to 'ready'
	stepStart = time.Now()
	logger.Info("step starting", "step_number", 7, "step_name", "update_status_ready")
	if err := p.db.UpdateVideoStatus(ctx, job.VideoID, "ready"); err != nil {
		return p.fail(ctx, logger, job.VideoID, "update_status_ready", pipelineStart, fmt.Errorf("failed to update status to ready: %w", err))
	}
	logger.Info("step completed", "step_number", 7, "step_name", "update_status_ready", "duration_ms", time.Since(stepStart).Milliseconds())

	totalDuration := time.Since(pipelineStart).Milliseconds()
	logger.Info("pipeline completed successfully",
		"total_duration_ms", totalDuration,
		"renditions_created", len(result.Renditions),
	)
	metrics.TranscodeDuration.Observe(time.Since(pipelineStart).Seconds())
	metrics.TranscodeJobsTotal.WithLabelValues("success").Inc()
	return nil
}

func (p *Pipeline) fail(ctx context.Context, logger *slog.Logger, videoID, failedStep string, pipelineStart time.Time, err error) error {
	logger.Error("pipeline failed",
		"failed_step", failedStep,
		"error", err.Error(),
		"duration_ms", time.Since(pipelineStart).Milliseconds(),
	)
	// Attempt to mark as failed in DB
	if updateErr := p.db.UpdateVideoStatus(ctx, videoID, "failed"); updateErr != nil {
		logger.Error("critical: also failed to update video status to failed", "update_error", updateErr.Error())
	}
	metrics.TranscodeJobsTotal.WithLabelValues("failed").Inc()
	return err
}

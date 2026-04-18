package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"streamforge/api/internal/cache"
	"streamforge/api/internal/config"
	"streamforge/api/internal/db"
	"streamforge/api/internal/middleware"
	"streamforge/api/internal/models"
	queuepkg "streamforge/api/internal/queue"
	storagepkg "streamforge/api/internal/storage"
)

type VideoHandler struct {
	cfg      config.Config
	db       *db.DB
	rabbitMQ *queuepkg.Publisher
	storage  *storagepkg.BlobStorage
	cache    *cache.Cache
}

func NewVideoHandler(cfg config.Config, db *db.DB, rmq *queuepkg.Publisher, blob *storagepkg.BlobStorage, cacheCli *cache.Cache) *VideoHandler {
	return &VideoHandler{
		cfg:      cfg,
		db:       db,
		rabbitMQ: rmq,
		storage:  blob,
		cache:    cacheCli,
	}
}

func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.List")

	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("starting", "user_id", userID)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	status := r.URL.Query().Get("status")

	// Cache key for this specific list state
	cacheKey := fmt.Sprintf("videos:list:%s:%d:%d:%s", userID, page, limit, status)

	if cachedVal, err := h.cache.GetVideoMeta(r.Context(), cacheKey); err == nil && cachedVal != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedVal))
		logger.Info("completed", "outcome", "success", "cache", "hit", "duration_ms", time.Since(start).Milliseconds())
		return
	}

	videos, total, err := h.db.ListVideosByUser(r.Context(), userID, page, limit)
	if err != nil {
		logger.Error("failed to list videos", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to list videos")
		return
	}

	var filtered []models.Video
	for _, v := range videos {
		if status != "" && string(v.Status) != status {
			continue
		}
		filtered = append(filtered, v)
	}

	if filtered == nil {
		filtered = []models.Video{}
	}

	resp := map[string]interface{}{
		"videos": filtered,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	_ = h.cache.SetVideoMeta(r.Context(), cacheKey, resp, 60*time.Second)

	respond(w, http.StatusOK, resp)
	logger.Info("completed", "outcome", "success", "total", total, "page", page, "duration_ms", time.Since(start).Milliseconds())
}

func (h *VideoHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.Search")

	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	searchQuery := r.URL.Query().Get("q")
	logger.Info("starting", "user_id", userID, "query", searchQuery)

	if searchQuery == "" {
		// If query is empty, fallback to List behavior
		h.List(w, r)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	videos, total, err := h.db.SearchVideosByUser(r.Context(), userID, searchQuery, page, limit)
	if err != nil {
		logger.Error("failed to search videos", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to search videos")
		return
	}

	if videos == nil {
		videos = []models.Video{}
	}

	resp := map[string]interface{}{
		"videos": videos,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	respond(w, http.StatusOK, resp)
	logger.Info("completed", "outcome", "success", "total", total, "duration_ms", time.Since(start).Milliseconds())
}

func (h *VideoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.GetByID")

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing video ID")
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)
	logger.Info("starting", "video_id", id, "user_id", userID)

	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		logger.Warn("video not found", "video_id", id, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	if video.UserID != userID && userRole != "admin" {
		logger.Warn("forbidden access attempt", "video_id", id, "owner_id", video.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	respond(w, http.StatusOK, video)
	logger.Info("completed", "outcome", "success", "video_id", id, "duration_ms", time.Since(start).Milliseconds())
}

func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.Upload")

	// Parse multipart form with max 2GB limit
	if err := r.ParseMultipartForm(2 << 30); err != nil {
		logger.Error("failed to parse multipart form", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, "Invalid multipart form or file too large (max 2GB)")
		return
	}
	defer r.MultipartForm.RemoveAll()

	// 1. Get userID from Context
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("starting", "user_id", userID)

	// 3. Read title and description
	title := r.FormValue("title")
	desc := r.FormValue("description")

	if title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// 4. Extract and validate video file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Video file is required")
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "video/") {
		respondError(w, http.StatusBadRequest, "File must be a valid video type (e.g. video/mp4)")
		return
	}

	logger.Info("upload file received",
		"file_size_bytes", fileHeader.Size,
		"content_type", contentType,
		"title", title,
	)

	// 5. Create video record
	videoID, err := h.db.CreateVideo(r.Context(), userID, title, desc)
	if err != nil {
		logger.Error("failed to create video record", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to register video in database")
		return
	}

	logger.Info("video record created", "video_id", videoID)

	// 6. Upload to blob
	blobURL, err := h.storage.UploadRawVideo(r.Context(), videoID, file, fileHeader.Size, contentType)
	if err != nil {
		logger.Error("blob upload failed", "video_id", videoID, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		// Update status to failed
		h.db.UpdateVideoStatus(r.Context(), videoID, string(models.StatusFailed))
		respondError(w, http.StatusInternalServerError, "Failed to upload video to storage")
		return
	}

	// 7. Update raw URL
	_ = h.db.UpdateVideoRawBlobURL(r.Context(), videoID, blobURL)

	// 8. Publish job
	err = h.rabbitMQ.PublishTranscodeJob(r.Context(), videoID, blobURL, userID)
	if err != nil {
		// Log error but gracefully continue, job can be manually requeued
		logger.Error("CRITICAL: failed to publish transcoding job to rabbitmq", "video_id", videoID, "error", err.Error())
	}

	// 9. Return
	respond(w, http.StatusCreated, map[string]string{
		"video_id": videoID,
		"status":   string(models.StatusPending),
	})
	logger.Info("completed", "outcome", "success",
		"video_id", videoID,
		"file_size_bytes", fileHeader.Size,
		"content_type", contentType,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// Delete removes a video entirely (raw blob, HLS blob folder, database)
func (h *VideoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.Delete")

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Video ID is required")
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("starting", "video_id", id, "user_id", userID)

	// Verify Ownership
	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		logger.Warn("video not found for deletion", "video_id", id, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	userRole := middleware.GetUserRole(r)
	if video.UserID != userID && userRole != "admin" {
		logger.Warn("forbidden delete attempt", "video_id", id, "owner_id", video.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusForbidden, "You do not have permission to delete this video")
		return
	}

	// Delete from Storage
	if err := h.storage.DeleteRawVideo(r.Context(), id); err != nil {
		logger.Error("failed to delete raw video blob", "video_id", id, "error", err.Error())
	}

	if err := h.storage.DeleteHLSFolder(r.Context(), id); err != nil {
		logger.Error("failed to delete HLS folder", "video_id", id, "error", err.Error())
	}

	// Delete from DB
	if err := h.db.DeleteVideo(r.Context(), id); err != nil {
		logger.Error("failed to delete video record", "video_id", id, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to delete video")
		return
	}

	// Invalidate Cache
	_ = h.cache.DeleteVideoMeta(r.Context(), id)
	_ = h.cache.InvalidateUserVideoList(r.Context(), userID)

	w.WriteHeader(http.StatusNoContent)
	logger.Info("completed", "outcome", "success", "video_id", id, "duration_ms", time.Since(start).Milliseconds())
}

func (h *VideoHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "VideoHandler.GetStreamURL")

	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Video ID is required")
		return
	}

	userID := middleware.GetUserID(r)
	logger.Info("starting", "video_id", id, "user_id", userID)

	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		logger.Warn("video not found", "video_id", id, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	if video.UserID != userID {
		logger.Warn("forbidden stream access", "video_id", id, "owner_id", video.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if video.Status != models.StatusReady {
		logger.Info("video not ready for streaming", "video_id", id, "video_status", string(video.Status), "duration_ms", time.Since(start).Milliseconds())
		respond(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "video is not ready yet",
			"status": string(video.Status),
		})
		return
	}

	// ↓ Only this block changes — everything above is untouched
	sasExpiry := time.Now().UTC().Add(2 * time.Hour)
	masterURL, sasToken, err := h.storage.GenerateVideoDirectorySASURL(r.Context(), id)
	if err != nil {
		logger.Error("SAS generation failed", "video_id", id, "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to generate stream url")
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"master_playlist_url": masterURL,
		"sas_token":           sasToken,
	})
	logger.Info("completed", "outcome", "success",
		"video_id", id,
		"sas_token_expiry", sasExpiry.Format(time.RFC3339),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

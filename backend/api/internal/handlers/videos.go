package handlers

import (
	"fmt"
	"log"
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
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
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

	status := r.URL.Query().Get("status")

	// Cache key for this specific list state
	cacheKey := fmt.Sprintf("videos:list:%s:%d:%d:%s", userID, page, limit, status)

	if cachedVal, err := h.cache.GetVideoMeta(r.Context(), cacheKey); err == nil && cachedVal != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedVal))
		return
	}

	videos, total, err := h.db.ListVideosByUser(r.Context(), userID, page, limit)
	if err != nil {
		log.Printf("list videos err: %v", err)
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
}

func (h *VideoHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	searchQuery := r.URL.Query().Get("q")
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
		log.Printf("search videos err: %v", err)
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
}

func (h *VideoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing video ID")
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	if video.UserID != userID && userRole != "admin" {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	respond(w, http.StatusOK, video)
}

func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max 2GB limit
	if err := r.ParseMultipartForm(2 << 30); err != nil {
		log.Printf("parse multipart err: %v", err)
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

	// 5. Create video record
	videoID, err := h.db.CreateVideo(r.Context(), userID, title, desc)
	if err != nil {
		log.Printf("db create video error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to register video in database")
		return
	}

	// 6. Upload to blob
	blobURL, err := h.storage.UploadRawVideo(r.Context(), videoID, file, fileHeader.Size, contentType)
	if err != nil {
		log.Printf("blob upload error: %v", err)
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
		log.Printf("CRITICAL: failed to publish transcoding job to rabbitmq: %v", err)
	}

	// 9. Return
	respond(w, http.StatusCreated, map[string]string{
		"video_id": videoID,
		"status":   string(models.StatusPending),
	})
}

// Delete removes a video entirely (raw blob, HLS blob folder, database)
func (h *VideoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	// Verify Ownership
	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	userRole := middleware.GetUserRole(r)
	if video.UserID != userID && userRole != "admin" {
		respondError(w, http.StatusForbidden, "You do not have permission to delete this video")
		return
	}

	// Delete from Storage
	if err := h.storage.DeleteRawVideo(r.Context(), id); err != nil {
		log.Printf("Failed to delete raw video blog %s: %v", id, err)
	}

	if err := h.storage.DeleteHLSFolder(r.Context(), id); err != nil {
		log.Printf("Failed to delete HLS folder %s: %v", id, err)
	}

	// Delete from DB
	if err := h.db.DeleteVideo(r.Context(), id); err != nil {
		log.Printf("Failed to delete video record %s: %v", id, err)
		respondError(w, http.StatusInternalServerError, "Failed to delete video")
		return
	}

	// Invalidate Cache
	_ = h.cache.DeleteVideoMeta(r.Context(), id)
	_ = h.cache.InvalidateUserVideoList(r.Context(), userID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Video ID is required")
		return
	}

	userID := middleware.GetUserID(r)

	video, err := h.db.GetVideoByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found")
		return
	}

	if video.UserID != userID {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if video.Status != models.StatusReady {
		respond(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "video is not ready yet",
			"status": string(video.Status),
		})
		return
	}

	// ↓ Only this block changes — everything above is untouched
	masterURL, sasToken, err := h.storage.GenerateVideoDirectorySASURL(r.Context(), id)
	// log.Printf("[Handler] generated URL: %s", masterURL)
	if err != nil {
		log.Printf("sas generation error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate stream url")
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"master_playlist_url": masterURL,
		"sas_token":           sasToken,
	})
}

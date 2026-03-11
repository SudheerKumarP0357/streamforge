package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"streamforge/api/internal/db"
	"streamforge/api/internal/middleware"
)

type WatchHandler struct {
	cosmos *db.CosmosDB
	db     *db.DB
}

func NewWatchHandler(cosmos *db.CosmosDB, pg *db.DB) *WatchHandler {
	return &WatchHandler{
		cosmos: cosmos,
		db:     pg,
	}
}

func (h *WatchHandler) SaveEvent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	videoID := chi.URLParam(r, "video_id")
	if videoID == "" {
		respondError(w, http.StatusBadRequest, "Missing video ID")
		return
	}

	var req struct {
		EventType       string `json:"event_type"`
		PositionSeconds int    `json:"position_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EventType != "play" && req.EventType != "pause" && req.EventType != "seek" && req.EventType != "end" {
		respondError(w, http.StatusBadRequest, "Invalid event type")
		return
	}

	event := db.WatchEvent{
		VideoID:         videoID,
		UserID:          userID,
		EventType:       req.EventType,
		Timestamp:       time.Now(),
		SessionID:       "", // Not explicitly tracked from UI currently
		PositionSeconds: req.PositionSeconds,
	}

	err := h.cosmos.SaveWatchEvent(r.Context(), event)
	if err != nil {
		log.Printf("Failed to save watch event: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to save event")
		return
	}

	if req.EventType == "end" {
		h.cosmos.UpsertWatchHistory(r.Context(), db.WatchHistory{
			UserID:          userID,
			VideoID:         videoID,
			WatchedAt:       time.Now(),
			ProgressSeconds: req.PositionSeconds,
			Completed:       true,
		})
	} else if req.EventType == "play" || req.EventType == "pause" {
		h.cosmos.UpsertWatchHistory(r.Context(), db.WatchHistory{
			UserID:          userID,
			VideoID:         videoID,
			WatchedAt:       time.Now(),
			ProgressSeconds: req.PositionSeconds,
			Completed:       false,
		})
	}

	respond(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

type HistoryResponse struct {
	VideoID         string             `json:"video_id"`
	Title           string             `json:"title"`
	WatchedAt       time.Time          `json:"watched_at"`
	ProgressSeconds int                `json:"progress_seconds"`
	DurationSeconds *int               `json:"duration_seconds"`
	Completed       bool               `json:"completed"`
}

func (h *WatchHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 20
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}

	history, err := h.cosmos.GetWatchHistory(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Failed to get watch history: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get history")
		return
	}

	var results []HistoryResponse
	for _, item := range history {
		v, err := h.db.GetVideoByID(r.Context(), item.VideoID)
		title := "Unknown Video"
		var duration *int
		if err == nil {
			title = v.Title
			duration = v.DurationSeconds
		}

		results = append(results, HistoryResponse{
			VideoID:         item.VideoID,
			Title:           title,
			WatchedAt:       item.WatchedAt,
			ProgressSeconds: item.ProgressSeconds,
			DurationSeconds: duration,
			Completed:       item.Completed,
		})
	}

	if results == nil {
		results = []HistoryResponse{}
	}

	respond(w, http.StatusOK, map[string]interface{}{"history": results})
}

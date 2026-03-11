package handlers

import (
	"net/http"
	"strconv"

	"streamforge/api/internal/db"
	"streamforge/api/internal/models"
)

type AdminHandler struct {
	db *db.DB
}

func NewAdminHandler(db *db.DB) *AdminHandler {
	return &AdminHandler{
		db: db,
	}
}

func (h *AdminHandler) GetVideos(w http.ResponseWriter, r *http.Request) {
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

	videos, total, err := h.db.ListAllVideosAdmin(r.Context(), page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list admin videos")
		return
	}

	if videos == nil {
		// Just to not return null in JSON
		videos = []models.AdminVideo{} // Wait, I need to import models here or use correct typings. Actually, let's fix imports first. 
	}

	resp := map[string]interface{}{
		"videos": videos,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	respond(w, http.StatusOK, resp)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetAdminStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get admin stats")
		return
	}

	respond(w, http.StatusOK, stats)
}

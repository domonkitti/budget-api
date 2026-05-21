package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsHandler struct {
	db *pgxpool.Pool
}

func NewSettingsHandler(db *pgxpool.Pool) *SettingsHandler {
	return &SettingsHandler{db: db}
}

func (h *SettingsHandler) GetActiveYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var value string
	err := h.db.QueryRow(ctx, `SELECT value FROM settings WHERE key = 'active_year'`).Scan(&value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		http.Error(w, "invalid active_year in settings", http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]int{"active_year": year})
}

func (h *SettingsHandler) SetActiveYear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		ActiveYear int `json:"active_year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.ActiveYear < 2500 || body.ActiveYear > 2700 {
		http.Error(w, "active_year out of range", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(ctx,
		`INSERT INTO settings (key, value) VALUES ('active_year', $1)
		 ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = now()`,
		strconv.Itoa(body.ActiveYear))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]int{"active_year": body.ActiveYear})
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotHandler struct {
	db *pgxpool.Pool
}

func NewSnapshotHandler(db *pgxpool.Pool) *SnapshotHandler {
	return &SnapshotHandler{db: db}
}

func (h *SnapshotHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, label, COALESCE(note, ''), created_at FROM snapshots ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []models.Snapshot{}
	for rows.Next() {
		var s models.Snapshot
		rows.Scan(&s.ID, &s.Label, &s.Note, &s.CreatedAt)
		result = append(result, s)
	}
	respond(w, http.StatusOK, result)
}

func (h *SnapshotHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
		Note  string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" {
		http.Error(w, "label required", http.StatusBadRequest)
		return
	}

	flat, err := queryFlat(r.Context(), h.db, map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataJSON, _ := json.Marshal(flat)

	var s models.Snapshot
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO snapshots (label, note, data) VALUES ($1, $2, $3)
		 RETURNING id, label, COALESCE(note, ''), created_at`,
		body.Label, body.Note, dataJSON).
		Scan(&s.ID, &s.Label, &s.Note, &s.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, s)
}

func (h *SnapshotHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sd models.SnapshotDetail
	var rawData []byte
	err := h.db.QueryRow(r.Context(),
		`SELECT id, label, COALESCE(note, ''), created_at, data FROM snapshots WHERE id = $1`, id).
		Scan(&sd.ID, &sd.Label, &sd.Note, &sd.CreatedAt, &rawData)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := json.Unmarshal(rawData, &sd.Data); err != nil {
		sd.Data = []models.FlatProject{}
	}
	respond(w, http.StatusOK, sd)
}

func (h *SnapshotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(r.Context(), `DELETE FROM snapshots WHERE id = $1`, id)
	w.WriteHeader(http.StatusNoContent)
}

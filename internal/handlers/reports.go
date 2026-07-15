package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportHandler struct {
	db *pgxpool.Pool
}

func NewReportHandler(db *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{db: db}
}

// -- Report groups --

func (h *ReportHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, name, sort_order FROM report_groups ORDER BY sort_order, id`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []models.ReportGroup{}
	for rows.Next() {
		var g models.ReportGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Order); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, g)
	}
	respond(w, http.StatusOK, result)
}

func (h *ReportHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	var g models.ReportGroup
	g.Name = body.Name
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO report_groups (name, sort_order)
		 VALUES ($1, COALESCE((SELECT MAX(sort_order) + 1 FROM report_groups), 0))
		 RETURNING id, sort_order`, body.Name).
		Scan(&g.ID, &g.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, g)
}

func (h *ReportHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if _, err := h.db.Exec(r.Context(), `UPDATE report_groups SET name = $1 WHERE id = $2`, body.Name, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReportHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(r.Context(), `DELETE FROM report_groups WHERE id = $1`, id)
	w.WriteHeader(http.StatusNoContent)
}

// -- Reports --

func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT id, group_id, preset_id, data FROM reports ORDER BY id`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []models.Report{}
	for rows.Next() {
		var rep models.Report
		var rawData []byte
		if err := rows.Scan(&rep.ID, &rep.GroupID, &rep.PresetID, &rawData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rep.Data = rawData
		result = append(result, rep)
	}
	respond(w, http.StatusOK, result)
}

func (h *ReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rep models.Report
	var rawData []byte
	err := h.db.QueryRow(r.Context(),
		`SELECT id, group_id, preset_id, data FROM reports WHERE id = $1`, id).
		Scan(&rep.ID, &rep.GroupID, &rep.PresetID, &rawData)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rep.Data = rawData
	respond(w, http.StatusOK, rep)
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GroupID int             `json:"groupId,string"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == 0 || len(body.Data) == 0 {
		http.Error(w, "groupId and data required", http.StatusBadRequest)
		return
	}
	var rep models.Report
	rep.GroupID = body.GroupID
	rep.Data = body.Data
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO reports (group_id, data) VALUES ($1, $2) RETURNING id, preset_id`,
		body.GroupID, []byte(body.Data)).
		Scan(&rep.ID, &rep.PresetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, rep)
}

// Update applies a partial patch — used both for the admin editor's debounced autosave (data
// only) and any future preset assignment (presetId only). Either field may be omitted.
func (h *ReportHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Data     *json.RawMessage `json:"data"`
		PresetID *string          `json:"presetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Data != nil {
		if _, err := h.db.Exec(r.Context(),
			`UPDATE reports SET data = $1, updated_at = NOW() WHERE id = $2`, []byte(*body.Data), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.PresetID != nil {
		if _, err := h.db.Exec(r.Context(), `UPDATE reports SET preset_id = $1 WHERE id = $2`, *body.PresetID, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(r.Context(), `DELETE FROM reports WHERE id = $1`, id)
	w.WriteHeader(http.StatusNoContent)
}

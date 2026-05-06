package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChangeLogHandler struct {
	db *pgxpool.Pool
}

func NewChangeLogHandler(db *pgxpool.Pool) *ChangeLogHandler {
	return &ChangeLogHandler{db: db}
}

func (h *ChangeLogHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	rows, err := h.db.Query(r.Context(), `
		SELECT cl.id, cl.table_name, cl.row_id, cl.project_id,
		       COALESCE(cl.row_name, ''), COALESCE(cl.fund_type, ''),
		       COALESCE(cl.data_year, 0), cl.field,
		       COALESCE(cl.old_value, 0), COALESCE(cl.new_value, 0), cl.changed_at,
		       cl.batch_id, cl.batch_comment
		FROM change_log cl
		JOIN projects p ON p.id = cl.project_id
		WHERE p.project_code = $1
		  AND cl.field IN ('budget', 'target', 'cut_transfer', 'under_budget')
		  AND NOT (cl.field = 'budget' AND cl.fund_type = 'ผูกพัน')
		ORDER BY cl.changed_at DESC, cl.id DESC
		LIMIT 20`, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []models.ChangeLogEntry{}
	for rows.Next() {
		var e models.ChangeLogEntry
		rows.Scan(&e.ID, &e.TableName, &e.RowID, &e.ProjectID,
			&e.RowName, &e.FundType, &e.DataYear, &e.Field,
			&e.OldValue, &e.NewValue, &e.ChangedAt,
			&e.BatchID, &e.BatchComment)
		result = append(result, e)
	}
	respond(w, http.StatusOK, result)
}

func (h *ChangeLogHandler) Undo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var e models.ChangeLogEntry
	err := h.db.QueryRow(r.Context(),
		`SELECT table_name, row_id, field, COALESCE(old_value, 0) FROM change_log WHERE id = $1`, id).
		Scan(&e.TableName, &e.RowID, &e.Field, &e.OldValue)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var sql string
	switch {
	case e.Field == "budget" && e.TableName == "sub_jobs":
		sql = `UPDATE sub_jobs SET budget = $1 WHERE id = $2`
	case e.Field == "target" && e.TableName == "sub_jobs":
		sql = `UPDATE sub_jobs SET target = $1 WHERE id = $2`
	case e.Field == "cut_transfer" && e.TableName == "sub_jobs":
		sql = `UPDATE sub_jobs SET cut_transfer = $1 WHERE id = $2`
	case e.Field == "under_budget" && e.TableName == "sub_jobs":
		sql = `UPDATE sub_jobs SET under_budget = $1 WHERE id = $2`
	case e.Field == "budget" && e.TableName == "budget_sources":
		sql = `UPDATE budget_sources SET budget = $1 WHERE id = $2`
	case e.Field == "target" && e.TableName == "budget_sources":
		sql = `UPDATE budget_sources SET target = $1 WHERE id = $2`
	case e.Field == "cut_transfer" && e.TableName == "budget_sources":
		sql = `UPDATE budget_sources SET cut_transfer = $1 WHERE id = $2`
	case e.Field == "under_budget" && e.TableName == "budget_sources":
		sql = `UPDATE budget_sources SET under_budget = $1 WHERE id = $2`
	default:
		http.Error(w, "unknown field", http.StatusBadRequest)
		return
	}

	if _, err := h.db.Exec(r.Context(), sql, e.OldValue, e.RowID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChangeLogHandler) UpdateBatchComment(w http.ResponseWriter, r *http.Request) {
	batchId := chi.URLParam(r, "batchId")
	var body struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if _, err := h.db.Exec(r.Context(),
		`UPDATE change_log SET batch_comment = $1 WHERE batch_id = $2`,
		body.Comment, batchId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

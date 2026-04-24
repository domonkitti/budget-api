package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectHandler struct {
	db *pgxpool.Pool
}

func NewProjectHandler(db *pgxpool.Pool) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year := q.Get("year")
	projectType := q.Get("type")
	division := q.Get("division")
	fundType := q.Get("fund_type")

	sql := `
		SELECT DISTINCT p.id, p.project_code, p.year, p.project_type,
		       p.item_no, p.name, p.division, p.department, p.created_at
		FROM projects p
		JOIN sub_jobs sj ON sj.project_id = p.id
		WHERE 1=1`
	args := []any{}
	i := 1

	if year != "" {
		sql += ` AND p.year = $` + strconv.Itoa(i)
		args = append(args, year)
		i++
	}
	if projectType != "" {
		sql += ` AND p.project_type = $` + strconv.Itoa(i)
		args = append(args, projectType)
		i++
	}
	if division != "" {
		sql += ` AND p.division = $` + strconv.Itoa(i)
		args = append(args, division)
		i++
	}
	if fundType != "" {
		sql += ` AND sj.fund_type = $` + strconv.Itoa(i)
		args = append(args, fundType)
		i++
	}
	sql += ` ORDER BY p.project_code`

	rows, err := h.db.Query(r.Context(), sql, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	projects := []models.Project{}
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.ProjectCode, &p.Year, &p.ProjectType,
			&p.ItemNo, &p.Name, &p.Division, &p.Department, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	respond(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var p models.Project
	err := h.db.QueryRow(r.Context(),
		`SELECT id, project_code, year, project_type, item_no, name, division, department, created_at
		 FROM projects WHERE project_code = $1`, code).
		Scan(&p.ID, &p.ProjectCode, &p.Year, &p.ProjectType,
			&p.ItemNo, &p.Name, &p.Division, &p.Department, &p.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	detail := models.ProjectDetail{Project: p}

	sjRows, _ := h.db.Query(r.Context(),
		`SELECT id, project_id, name, sort_order, fund_type, budget, target, budget-target
		 FROM sub_jobs WHERE project_id = $1 ORDER BY sort_order`, p.ID)
	defer sjRows.Close()
	for sjRows.Next() {
		var sj models.SubJob
		sjRows.Scan(&sj.ID, &sj.ProjectID, &sj.Name, &sj.SortOrder,
			&sj.FundType, &sj.Budget, &sj.Target, &sj.Remain)
		detail.SubJobs = append(detail.SubJobs, sj)
	}

	bsRows, _ := h.db.Query(r.Context(),
		`SELECT id, project_id, source, fund_type, budget, target, budget-target
		 FROM budget_sources WHERE project_id = $1 ORDER BY source, fund_type`, p.ID)
	defer bsRows.Close()
	for bsRows.Next() {
		var bs models.BudgetSource
		bsRows.Scan(&bs.ID, &bs.ProjectID, &bs.Source, &bs.FundType,
			&bs.Budget, &bs.Target, &bs.Remain)
		detail.BudgetSources = append(detail.BudgetSources, bs)
	}

	respond(w, http.StatusOK, detail)
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

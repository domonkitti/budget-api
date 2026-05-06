package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScenarioHandler struct {
	db *pgxpool.Pool
}

func NewScenarioHandler(db *pgxpool.Pool) *ScenarioHandler {
	return &ScenarioHandler{db: db}
}

func (h *ScenarioHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, label, COALESCE(note, ''), created_at, updated_at FROM scenarios ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []models.Scenario{}
	for rows.Next() {
		var s models.Scenario
		rows.Scan(&s.ID, &s.Label, &s.Note, &s.CreatedAt, &s.UpdatedAt)
		result = append(result, s)
	}
	respond(w, http.StatusOK, result)
}

func (h *ScenarioHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
		Note  string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" {
		http.Error(w, "label required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var s models.Scenario
	err = tx.QueryRow(ctx,
		`INSERT INTO scenarios (label, note) VALUES ($1, $2)
		 RETURNING id, label, COALESCE(note, ''), created_at, updated_at`,
		body.Label, body.Note).
		Scan(&s.ID, &s.Label, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO scenario_sub_jobs (scenario_id, project_id, name, sort_order, fund_type, data_year, budget, target, cut_transfer, under_budget)
		SELECT $1, project_id, name, MIN(sort_order), fund_type, data_year,
		       SUM(budget), SUM(target), SUM(cut_transfer), SUM(under_budget)
		FROM sub_jobs
		GROUP BY project_id, name, fund_type, data_year`, s.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO scenario_budget_sources (scenario_id, project_id, source, fund_type, data_year, budget, target, cut_transfer, under_budget)
		SELECT $1, project_id, source, fund_type, data_year,
		       SUM(budget), SUM(target), SUM(cut_transfer), SUM(under_budget)
		FROM budget_sources
		GROUP BY project_id, source, fund_type, data_year`, s.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, s)
}

func (h *ScenarioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(r.Context(), `DELETE FROM scenarios WHERE id = $1`, id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScenarioHandler) Flat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := queryScenarioFlat(r.Context(), h.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, result)
}

func (h *ScenarioHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	scenID := chi.URLParam(r, "id")
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

	sjRows, err := h.db.Query(r.Context(),
		`SELECT id, project_id, name, sort_order, fund_type, data_year,
		        budget, target, budget - target AS remain, cut_transfer, under_budget
		 FROM scenario_sub_jobs
		 WHERE scenario_id = $1 AND project_id = $2
		 ORDER BY sort_order, name, data_year, fund_type, id`,
		scenID, p.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer sjRows.Close()
	subJobs := []models.SubJob{}
	for sjRows.Next() {
		var sj models.SubJob
		sjRows.Scan(&sj.ID, &sj.ProjectID, &sj.Name, &sj.SortOrder,
			&sj.FundType, &sj.DataYear, &sj.Budget, &sj.Target, &sj.Remain,
			&sj.CutTransfer, &sj.UnderBudget)
		subJobs = append(subJobs, sj)
	}

	bsRows, err := h.db.Query(r.Context(),
		`SELECT id, project_id, source, fund_type, data_year,
		        budget, target, budget - target AS remain, cut_transfer, under_budget
		 FROM scenario_budget_sources
		 WHERE scenario_id = $1 AND project_id = $2
		 ORDER BY source, data_year, fund_type, id`,
		scenID, p.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer bsRows.Close()
	budgetSources := []models.BudgetSource{}
	for bsRows.Next() {
		var bs models.BudgetSource
		bsRows.Scan(&bs.ID, &bs.ProjectID, &bs.Source, &bs.FundType,
			&bs.DataYear, &bs.Budget, &bs.Target, &bs.Remain, &bs.CutTransfer, &bs.UnderBudget)
		budgetSources = append(budgetSources, bs)
	}

	respond(w, http.StatusOK, models.ProjectDetail{
		Project:       p,
		SubJobs:       subJobs,
		BudgetSources: budgetSources,
	})
}

func (h *ScenarioHandler) UpdateSubJob(w http.ResponseWriter, r *http.Request) {
	sjID := chi.URLParam(r, "sjID")
	var body struct {
		Budget      float64 `json:"budget"`
		Target      float64 `json:"target"`
		CutTransfer float64 `json:"cut_transfer"`
		UnderBudget float64 `json:"under_budget"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if _, err := h.db.Exec(r.Context(),
		`UPDATE scenario_sub_jobs SET budget = $1, target = $2, cut_transfer = $3, under_budget = $4 WHERE id = $5`,
		body.Budget, body.Target, body.CutTransfer, body.UnderBudget, sjID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScenarioHandler) Promote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `
		UPDATE sub_jobs sj
		SET budget = ssj.budget, target = ssj.target, cut_transfer = ssj.cut_transfer, under_budget = ssj.under_budget
		FROM scenario_sub_jobs ssj
		WHERE ssj.scenario_id = $1
		  AND ssj.project_id = sj.project_id
		  AND ssj.name      = sj.name
		  AND ssj.fund_type = sj.fund_type
		  AND ssj.data_year = sj.data_year`, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec(ctx, `
		UPDATE budget_sources bs
		SET budget = sbs.budget, target = sbs.target, cut_transfer = sbs.cut_transfer, under_budget = sbs.under_budget
		FROM scenario_budget_sources sbs
		WHERE sbs.scenario_id = $1
		  AND sbs.project_id = bs.project_id
		  AND sbs.source    = bs.source
		  AND sbs.fund_type = bs.fund_type
		  AND sbs.data_year = bs.data_year`, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScenarioHandler) UpdateBudgetSource(w http.ResponseWriter, r *http.Request) {
	bsID := chi.URLParam(r, "bsID")
	var body struct {
		Budget      float64 `json:"budget"`
		Target      float64 `json:"target"`
		CutTransfer float64 `json:"cut_transfer"`
		UnderBudget float64 `json:"under_budget"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if _, err := h.db.Exec(r.Context(),
		`UPDATE scenario_budget_sources SET budget = $1, target = $2, cut_transfer = $3, under_budget = $4 WHERE id = $5`,
		body.Budget, body.Target, body.CutTransfer, body.UnderBudget, bsID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func queryScenarioFlat(ctx context.Context, db *pgxpool.Pool, scenID string) ([]models.FlatProject, error) {
	sql := `
		SELECT
			p.id, p.project_code, p.item_no, p.name, p.division, p.project_type, p.year,
			COALESCE(
				(SELECT json_agg(row_data ORDER BY (row_data->>'year')::int, row_data->>'source', row_data->>'fund_type')
				 FROM (
					SELECT json_build_object(
						'year',      data_year,
						'source',    source,
						'fund_type', fund_type,
						'budget',    SUM(budget),
						'target',    SUM(target),
						'remain',    SUM(budget - target)
					) AS row_data
					FROM scenario_budget_sources sbs
					WHERE sbs.project_id = p.id AND sbs.scenario_id = $1
					GROUP BY data_year, source, fund_type
				 ) sub),
				'[]'::json
			) AS source_breakdown,
			COALESCE(
				(SELECT json_agg(row_data ORDER BY COALESCE((row_data->>'sort_order')::int, 999999), row_data->>'name', (row_data->>'year')::int, row_data->>'fund_type')
				 FROM (
					SELECT json_build_object(
						'name',       name,
						'sort_order', sort_order,
						'year',       data_year,
						'fund_type',  fund_type,
						'budget',     SUM(budget),
						'target',     SUM(target),
						'remain',     SUM(budget - target)
					) AS row_data
					FROM scenario_sub_jobs ssj
					WHERE ssj.project_id = p.id AND ssj.scenario_id = $1
					GROUP BY name, sort_order, data_year, fund_type
				 ) sub),
				'[]'::json
			) AS sub_jobs
		FROM projects p
		WHERE p.id IN (
			SELECT DISTINCT project_id FROM scenario_sub_jobs WHERE scenario_id = $1
			UNION
			SELECT DISTINCT project_id FROM scenario_budget_sources WHERE scenario_id = $1
		)
		ORDER BY p.project_code`

	rows, err := db.Query(ctx, sql, scenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []models.FlatProject{}
	for rows.Next() {
		var fp models.FlatProject
		var rawBreakdown []byte
		var rawSubJobs []byte
		if err := rows.Scan(
			&fp.ID, &fp.ProjectCode, &fp.ItemNo, &fp.Name, &fp.Division, &fp.ProjectType, &fp.Year,
			&rawBreakdown, &rawSubJobs,
		); err != nil {
			return nil, err
		}
		if len(rawBreakdown) > 0 {
			_ = json.Unmarshal(rawBreakdown, &fp.SourceBreakdown)
		}
		if fp.SourceBreakdown == nil {
			fp.SourceBreakdown = []models.SourceYearEntry{}
		}
		if len(rawSubJobs) > 0 {
			_ = json.Unmarshal(rawSubJobs, &fp.SubJobs)
		}
		if fp.SubJobs == nil {
			fp.SubJobs = []models.SubJobYearEntry{}
		}
		result = append(result, fp)
	}
	return result, nil
}

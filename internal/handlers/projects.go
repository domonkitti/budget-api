package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		       p.item_no, p.name, p.division, p.department, p.project_group, p.created_at
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
			&p.ItemNo, &p.Name, &p.Division, &p.Department, &p.GroupName, &p.CreatedAt); err != nil {
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
		`SELECT id, project_code, year, project_type, item_no, name, division, department, project_group, created_at
		 FROM projects WHERE project_code = $1`, code).
		Scan(&p.ID, &p.ProjectCode, &p.Year, &p.ProjectType,
			&p.ItemNo, &p.Name, &p.Division, &p.Department, &p.GroupName, &p.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	detail := models.ProjectDetail{
		Project:       p,
		SubJobs:       []models.SubJob{},
		BudgetSources: []models.BudgetSource{},
	}

	sjRows, _ := h.db.Query(r.Context(),
		`SELECT MIN(id), project_id, name, MIN(sort_order), fund_type, data_year,
		        SUM(budget), SUM(target), SUM(budget)-SUM(target),
		        SUM(cut_transfer), SUM(under_budget)
		 FROM sub_jobs WHERE project_id = $1
		 GROUP BY project_id, name, fund_type, data_year
		 ORDER BY MIN(sort_order), name, data_year, fund_type`, p.ID)
	defer sjRows.Close()
	for sjRows.Next() {
		var sj models.SubJob
		sjRows.Scan(&sj.ID, &sj.ProjectID, &sj.Name, &sj.SortOrder,
			&sj.FundType, &sj.DataYear, &sj.Budget, &sj.Target, &sj.Remain,
			&sj.CutTransfer, &sj.UnderBudget)
		detail.SubJobs = append(detail.SubJobs, sj)
	}

	bsRows, _ := h.db.Query(r.Context(),
		`SELECT MIN(id), project_id, source, fund_type, data_year,
		        SUM(budget), SUM(target), SUM(budget)-SUM(target),
		        SUM(cut_transfer), SUM(under_budget)
		 FROM budget_sources WHERE project_id = $1
		 GROUP BY project_id, source, fund_type, data_year
		 ORDER BY source, data_year, fund_type`, p.ID)
	defer bsRows.Close()
	for bsRows.Next() {
		var bs models.BudgetSource
		bsRows.Scan(&bs.ID, &bs.ProjectID, &bs.Source, &bs.FundType,
			&bs.DataYear, &bs.Budget, &bs.Target, &bs.Remain, &bs.CutTransfer, &bs.UnderBudget)
		detail.BudgetSources = append(detail.BudgetSources, bs)
	}

	respond(w, http.StatusOK, detail)
}

func (h *ProjectHandler) Flat(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := queryFlat(r.Context(), h.db, map[string]string{
		"year":        q.Get("year"),
		"years":       q.Get("years"),
		"type":        q.Get("type"),
		"division":    q.Get("division"),
		"source":      q.Get("source"),
		"active_only": q.Get("active_only"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, result)
}

func (h *ProjectHandler) CreateSubJob(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProjectID int     `json:"project_id"`
		Name      string  `json:"name"`
		SortOrder *int    `json:"sort_order"`
		FundType  string  `json:"fund_type"`
		DataYear  int     `json:"data_year"`
		Budget    float64 `json:"budget"`
		Target    float64 `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var sj models.SubJob
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO sub_jobs (project_id, name, sort_order, fund_type, data_year, budget, target)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, project_id, name, sort_order, fund_type, data_year, budget, target, budget-target`,
		body.ProjectID, body.Name, body.SortOrder, body.FundType, body.DataYear, body.Budget, body.Target,
	).Scan(&sj.ID, &sj.ProjectID, &sj.Name, &sj.SortOrder, &sj.FundType, &sj.DataYear, &sj.Budget, &sj.Target, &sj.Remain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, sj)
}

func (h *ProjectHandler) CreateBudgetSource(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProjectID   int     `json:"project_id"`
		Source      string  `json:"source"`
		FundType    string  `json:"fund_type"`
		DataYear    int     `json:"data_year"`
		Budget      float64 `json:"budget"`
		Target      float64 `json:"target"`
		CutTransfer float64 `json:"cut_transfer"`
		UnderBudget float64 `json:"under_budget"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var bs models.BudgetSource
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO budget_sources (project_id, source, fund_type, data_year, budget, target, cut_transfer, under_budget)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, project_id, source, fund_type, data_year, budget, target, budget-target, cut_transfer, under_budget`,
		body.ProjectID, body.Source, body.FundType, body.DataYear, body.Budget, body.Target, body.CutTransfer, body.UnderBudget,
	).Scan(&bs.ID, &bs.ProjectID, &bs.Source, &bs.FundType, &bs.DataYear, &bs.Budget, &bs.Target, &bs.Remain, &bs.CutTransfer, &bs.UnderBudget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusCreated, bs)
}

func (h *ProjectHandler) UpdateInfo(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	var body struct {
		Name        string  `json:"name"`
		ItemNo      *string `json:"item_no"`
		Year        int     `json:"year"`
		ProjectType string  `json:"project_type"`
		Division    *string `json:"division"`
		Department  *string `json:"department"`
		GroupName   *string `json:"group_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(r.Context(),
		`UPDATE projects SET name=$1, item_no=$2, year=$3, project_type=$4, division=$5, department=$6, project_group=$7
		 WHERE project_code=$8`,
		body.Name, body.ItemNo, body.Year, body.ProjectType, body.Division, body.Department, body.GroupName, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (h *ProjectHandler) UpdateSubJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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
		`UPDATE sub_jobs SET budget = $1, target = $2 WHERE id = $3`,
		body.Budget, body.Target, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) UpdateBudgetSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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
		`UPDATE budget_sources SET budget = $1, target = $2, cut_transfer = $3, under_budget = $4 WHERE id = $5`,
		body.Budget, body.Target, body.CutTransfer, body.UnderBudget, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) BatchSave(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BatchID       string `json:"batch_id"`
		BatchComment  string `json:"batch_comment"`
		SubJobUpdates []struct {
			ID          int     `json:"id"`
			Budget      float64 `json:"budget"`
			Target      float64 `json:"target"`
			CutTransfer float64 `json:"cut_transfer"`
			UnderBudget float64 `json:"under_budget"`
		} `json:"sub_job_updates"`
		BudgetSourceUpdates []struct {
			ID          int     `json:"id"`
			Budget      float64 `json:"budget"`
			Target      float64 `json:"target"`
			CutTransfer float64 `json:"cut_transfer"`
			UnderBudget float64 `json:"under_budget"`
		} `json:"budget_source_updates"`
		NewSubJobs []struct {
			ProjectID   int     `json:"project_id"`
			Name        string  `json:"name"`
			SortOrder   *int    `json:"sort_order"`
			FundType    string  `json:"fund_type"`
			DataYear    int     `json:"data_year"`
			Budget      float64 `json:"budget"`
			Target      float64 `json:"target"`
			CutTransfer float64 `json:"cut_transfer"`
			UnderBudget float64 `json:"under_budget"`
		} `json:"new_sub_jobs"`
		NewBudgetSources []struct {
			ProjectID   int     `json:"project_id"`
			Source      string  `json:"source"`
			FundType    string  `json:"fund_type"`
			DataYear    int     `json:"data_year"`
			Budget      float64 `json:"budget"`
			Target      float64 `json:"target"`
			CutTransfer float64 `json:"cut_transfer"`
			UnderBudget float64 `json:"under_budget"`
		} `json:"new_budget_sources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	if body.BatchID != "" {
		if _, err := tx.Exec(r.Context(), `SELECT set_config('app.batch_id', $1, true)`, body.BatchID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, sj := range body.SubJobUpdates {
		if _, err := tx.Exec(r.Context(),
			`UPDATE sub_jobs SET budget = $1, target = $2, cut_transfer = $3, under_budget = $4 WHERE id = $5`,
			sj.Budget, sj.Target, sj.CutTransfer, sj.UnderBudget, sj.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec(r.Context(), `
			DELETE FROM sub_jobs
			WHERE project_id = (SELECT project_id FROM sub_jobs WHERE id = $1)
			  AND name       = (SELECT name       FROM sub_jobs WHERE id = $1)
			  AND fund_type  = (SELECT fund_type  FROM sub_jobs WHERE id = $1)
			  AND data_year  = (SELECT data_year  FROM sub_jobs WHERE id = $1)
			  AND id != $1`, sj.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, bs := range body.BudgetSourceUpdates {
		if _, err := tx.Exec(r.Context(),
			`UPDATE budget_sources SET budget = $1, target = $2, cut_transfer = $3, under_budget = $4 WHERE id = $5`,
			bs.Budget, bs.Target, bs.CutTransfer, bs.UnderBudget, bs.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec(r.Context(), `
			DELETE FROM budget_sources
			WHERE project_id = (SELECT project_id FROM budget_sources WHERE id = $1)
			  AND source     = (SELECT source     FROM budget_sources WHERE id = $1)
			  AND fund_type  = (SELECT fund_type  FROM budget_sources WHERE id = $1)
			  AND data_year  = (SELECT data_year  FROM budget_sources WHERE id = $1)
			  AND id != $1`, bs.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, sj := range body.NewSubJobs {
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO sub_jobs (project_id, name, sort_order, fund_type, data_year, budget, target, cut_transfer, under_budget) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			sj.ProjectID, sj.Name, sj.SortOrder, sj.FundType, sj.DataYear, sj.Budget, sj.Target, sj.CutTransfer, sj.UnderBudget); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, bs := range body.NewBudgetSources {
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO budget_sources (project_id, source, fund_type, data_year, budget, target, cut_transfer, under_budget) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			bs.ProjectID, bs.Source, bs.FundType, bs.DataYear, bs.Budget, bs.Target, bs.CutTransfer, bs.UnderBudget); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if body.BatchID != "" && body.BatchComment != "" {
		if _, err := tx.Exec(r.Context(),
			`UPDATE change_log SET batch_comment = $1 WHERE batch_id = $2 AND batch_comment = ''`,
			body.BatchComment, body.BatchID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := pruneChangeLog(r.Context(), tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type changeLogPruner interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func pruneChangeLog(ctx context.Context, execer changeLogPruner) error {
	_, err := execer.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY changed_at DESC, id DESC) AS rn
			FROM change_log
		)
		DELETE FROM change_log
		USING ranked
		WHERE change_log.id = ranked.id
		  AND ranked.rn > 20`)
	return err
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// queryFlat is the shared flat-project query used by both the Flat handler and snapshot creation.
func queryFlat(ctx context.Context, db *pgxpool.Pool, params map[string]string) ([]models.FlatProject, error) {
	year := params["year"]
	yearsParam := params["years"]
	projectType := params["type"]
	division := params["division"]
	department := params["department"]
	groupName := params["group"]
	source := params["source"]
	activeOnly := params["active_only"] == "true" || params["active_only"] == "1"
	years := parseYearsParam(yearsParam)
	if len(years) == 0 && year != "" {
		if parsed, err := strconv.Atoi(year); err == nil {
			years = []int{parsed}
		}
	}

	sql := `
		SELECT
			p.id, p.project_code, p.item_no, p.name, p.division, p.department, p.project_group, p.project_type, p.year,
			COALESCE(
				(SELECT json_agg(row_data ORDER BY (row_data->>'year')::int, row_data->>'source', row_data->>'fund_type')
				 FROM (
					SELECT json_build_object(
						'year',         data_year,
						'source',       source,
						'fund_type',    fund_type,
						'budget',       SUM(budget),
						'target',       SUM(target),
						'remain',       SUM(budget - target),
						'cut_transfer', SUM(cut_transfer),
						'under_budget', SUM(under_budget)
					) AS row_data
					FROM budget_sources bs
					WHERE bs.project_id = p.id
					  AND ($1::int[] IS NULL OR bs.data_year = ANY($1::int[]))
					GROUP BY data_year, source, fund_type
				 ) sub),
				'[]'::json
			) AS source_breakdown,
			COALESCE(
				(SELECT json_agg(row_data ORDER BY COALESCE((row_data->>'sort_order')::int, 999999), row_data->>'name', (row_data->>'year')::int, row_data->>'fund_type')
				 FROM (
					SELECT json_build_object(
						'name',         name,
						'sort_order',   sort_order,
						'year',         data_year,
						'fund_type',    fund_type,
						'budget',       SUM(budget),
						'target',       SUM(target),
						'remain',       SUM(budget - target),
						'cut_transfer', SUM(cut_transfer),
						'under_budget', SUM(under_budget)
					) AS row_data
					FROM sub_jobs sj
					WHERE sj.project_id = p.id
					  AND ($1::int[] IS NULL OR sj.data_year = ANY($1::int[]))
					GROUP BY name, sort_order, data_year, fund_type
				 ) sub),
				'[]'::json
			) AS sub_jobs
		FROM projects p
		WHERE 1=1`
	args := []any{nil}
	if len(years) > 0 {
		args[0] = years
	}
	i := 2

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
	if department != "" {
		sql += ` AND p.department = $` + strconv.Itoa(i)
		args = append(args, department)
		i++
	}
	if groupName != "" {
		sql += ` AND p.project_group = $` + strconv.Itoa(i)
		args = append(args, groupName)
		i++
	}
	if source != "" {
		sql += ` AND p.id IN (SELECT project_id FROM budget_sources WHERE source = $` + strconv.Itoa(i) + `)`
		args = append(args, source)
		i++
	}
	if activeOnly && len(years) > 0 {
		sql += ` AND EXISTS (
			SELECT 1
			FROM budget_sources active_bs
			WHERE active_bs.project_id = p.id
			  AND active_bs.data_year = ANY($1::int[])
			GROUP BY active_bs.project_id
			HAVING SUM(active_bs.budget) > 0
		)`
	}
	sql += ` ORDER BY
		CASE p.project_type WHEN 'Y' THEN 1 WHEN 'C' THEN 2 WHEN 'L' THEN 3 ELSE 4 END,
		CASE p.project_group
			WHEN 'หมวดสิ่งก่อสร้าง' THEN 1
			WHEN 'หมวดเครื่องจักรอุปกรณ์' THEN 2
			WHEN 'หมวดเครื่องใช้สำนักงานและเครื่องมือเครื่องใช้ขนาดเล็ก' THEN 3
			WHEN 'หมวดวิจัยและพัฒนา' THEN 4
			WHEN 'หมวดลงทุนอื่นๆ' THEN 5
			WHEN 'หมวดสำรองราคา' THEN 6
			WHEN 'หมวดสำรองกรณีจำเป็นเร่งด่วน' THEN 7
			ELSE 99
		END,
		p.year, p.project_code`

	rows, err := db.Query(ctx, sql, args...)
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
			&fp.ID, &fp.ProjectCode, &fp.ItemNo, &fp.Name, &fp.Division, &fp.Department, &fp.GroupName, &fp.ProjectType, &fp.Year,
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

func parseYearsParam(value string) []int {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	years := make([]int, 0, len(parts))
	for _, part := range parts {
		year, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			years = append(years, year)
		}
	}
	return years
}

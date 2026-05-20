package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/domonkitti/budget-app-api/internal/po"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImportHandler struct {
	db       *pgxpool.Pool
	poClient po.Client
}

func NewImportHandler(db *pgxpool.Pool, poClient po.Client) *ImportHandler {
	return &ImportHandler{db: db, poClient: poClient}
}

// CheckVersions fetches PO versions, syncs import_status, returns list with status badges.
func (h *ImportHandler) CheckVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	poVersions, err := h.poClient.FetchVersions(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch po versions: %v", err), http.StatusBadGateway)
		return
	}

	codes := make([]string, len(poVersions))
	for i, v := range poVersions {
		codes[i] = v.ProjectCode
	}

	rows, err := h.db.Query(ctx,
		`SELECT project_code, last_accepted_version, last_accepted_at, po_version, po_updated_at, status
		 FROM po_import_status WHERE project_code = ANY($1)`, codes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	existing := map[string]models.ImportStatus{}
	for rows.Next() {
		var s models.ImportStatus
		if err := rows.Scan(&s.ProjectCode, &s.LastAcceptedVersion, &s.LastAcceptedAt,
			&s.POVersion, &s.POUpdatedAt, &s.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existing[s.ProjectCode] = s
	}

	result := make([]models.ImportStatus, 0, len(poVersions))
	for _, v := range poVersions {
		ver := v.Version
		updAt := v.UpdatedAt

		s, found := existing[v.ProjectCode]
		if !found {
			s = models.ImportStatus{
				ProjectCode: v.ProjectCode,
				POVersion:   &ver,
				POUpdatedAt: &updAt,
				Status:      "new",
			}
		} else {
			s.POVersion = &ver
			s.POUpdatedAt = &updAt
			if s.LastAcceptedVersion == nil || *s.LastAcceptedVersion < v.Version {
				s.Status = "has_update"
			} else {
				s.Status = "up_to_date"
			}
		}

		h.db.Exec(ctx, `
			INSERT INTO po_import_status (project_code, po_version, po_updated_at, status)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (project_code) DO UPDATE
			SET po_version=$2, po_updated_at=$3, status=$4`,
			s.ProjectCode, ver, updAt, s.Status)

		result = append(result, s)
	}

	respond(w, http.StatusOK, result)
}

// FetchDiff pulls the project from PO and diffs it against BG DB.
func (h *ImportHandler) FetchDiff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := chi.URLParam(r, "code")

	poPrj, err := h.poClient.FetchProject(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch po project: %v", err), http.StatusBadGateway)
		return
	}

	var bgPrj models.Project
	err = h.db.QueryRow(ctx,
		`SELECT id, project_code, year, project_type, item_no, name, division, department, project_group, created_at
		 FROM projects WHERE project_code=$1`, code).
		Scan(&bgPrj.ID, &bgPrj.ProjectCode, &bgPrj.Year, &bgPrj.ProjectType,
			&bgPrj.ItemNo, &bgPrj.Name, &bgPrj.Division, &bgPrj.Department, &bgPrj.GroupName, &bgPrj.CreatedAt)
	if err != nil {
		// project doesn't exist in BG yet — everything is "added"
		respond(w, http.StatusOK, models.ProjectDiff{
			ProjectCode:  code,
			POVersion:    poPrj.Version,
			HasChanges:   true,
			ProjectDiffs: []models.FieldDiff{{Field: "project", BGValue: nil, POValue: "new project"}},
			SubJobDiffs:  subJobDiffsAllAdded(poPrj.SubJobs),
		})
		return
	}

	sjRows, err := h.db.Query(ctx,
		`SELECT name, fund_type, data_year, budget, target FROM sub_jobs WHERE project_id=$1`, bgPrj.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer sjRows.Close()

	type sjKey struct {
		name, fundType string
		year           int
	}
	bgSJs := map[sjKey]models.POSubJob{}
	for sjRows.Next() {
		var sj models.POSubJob
		if err := sjRows.Scan(&sj.Name, &sj.FundType, &sj.DataYear, &sj.Budget, &sj.Target); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bgSJs[sjKey{sj.Name, sj.FundType, sj.DataYear}] = sj
	}

	diff := models.ProjectDiff{
		ProjectCode:  code,
		POVersion:    poPrj.Version,
		ProjectDiffs: diffProjectFields(bgPrj, *poPrj),
	}

	poKeys := map[sjKey]bool{}
	for _, sj := range poPrj.SubJobs {
		k := sjKey{sj.Name, sj.FundType, sj.DataYear}
		poKeys[k] = true
		if bgSJ, exists := bgSJs[k]; !exists {
			diff.SubJobDiffs = append(diff.SubJobDiffs, models.SubJobDiff{
				Name: sj.Name, FundType: sj.FundType, DataYear: sj.DataYear,
				Change: "added",
				Diffs:  []models.FieldDiff{{Field: "budget", BGValue: nil, POValue: sj.Budget}, {Field: "target", BGValue: nil, POValue: sj.Target}},
			})
		} else {
			fieldDiffs := diffSubJobFields(bgSJ, sj)
			change := "unchanged"
			if len(fieldDiffs) > 0 {
				change = "modified"
			}
			diff.SubJobDiffs = append(diff.SubJobDiffs, models.SubJobDiff{
				Name: sj.Name, FundType: sj.FundType, DataYear: sj.DataYear,
				Change: change, Diffs: fieldDiffs,
			})
		}
	}
	for k, bgSJ := range bgSJs {
		if !poKeys[k] {
			diff.SubJobDiffs = append(diff.SubJobDiffs, models.SubJobDiff{
				Name: bgSJ.Name, FundType: bgSJ.FundType, DataYear: bgSJ.DataYear,
				Change: "removed",
			})
		}
	}

	diff.HasChanges = len(diff.ProjectDiffs) > 0 || hasChangedSubJobs(diff.SubJobDiffs)
	respond(w, http.StatusOK, diff)
}

// Accept applies PO project data to BG DB and logs it.
func (h *ImportHandler) Accept(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := chi.URLParam(r, "code")

	poPrj, err := h.poClient.FetchProject(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch po project: %v", err), http.StatusBadGateway)
		return
	}

	if err := h.applyAccept(ctx, poPrj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, http.StatusOK, map[string]any{"ok": true, "project_code": code, "po_version": poPrj.Version})
}

// BatchAccept accepts multiple projects in one request.
func (h *ImportHandler) BatchAccept(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectCodes []string `json:"project_codes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	results := make([]map[string]any, 0, len(req.ProjectCodes))
	for _, code := range req.ProjectCodes {
		poPrj, err := h.poClient.FetchProject(r.Context(), code)
		if err != nil {
			results = append(results, map[string]any{"project_code": code, "ok": false, "error": err.Error()})
			continue
		}
		if err := h.applyAccept(r.Context(), poPrj); err != nil {
			results = append(results, map[string]any{"project_code": code, "ok": false, "error": err.Error()})
		} else {
			results = append(results, map[string]any{"project_code": code, "ok": true, "po_version": poPrj.Version})
		}
	}
	respond(w, http.StatusOK, map[string]any{"results": results})
}

// ListLog returns import audit log, optionally filtered by project_code.
func (h *ImportHandler) ListLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("project_code")

	sql := `SELECT id, project_code, po_version, accepted_by, accepted_at, snapshot_json
	        FROM po_import_log`
	args := []any{}
	if code != "" {
		sql += ` WHERE project_code=$1`
		args = append(args, code)
	}
	sql += ` ORDER BY accepted_at DESC LIMIT 100`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	logs := []models.ImportLog{}
	for rows.Next() {
		var l models.ImportLog
		if err := rows.Scan(&l.ID, &l.ProjectCode, &l.POVersion, &l.AcceptedBy, &l.AcceptedAt, &l.SnapshotJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}
	respond(w, http.StatusOK, logs)
}

// applyAccept runs the full accept transaction for one project.
func (h *ImportHandler) applyAccept(ctx context.Context, poPrj *models.POProject) error {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var projectID int
	if err := tx.QueryRow(ctx, `SELECT id FROM projects WHERE project_code=$1`, poPrj.ProjectCode).Scan(&projectID); err != nil {
		return fmt.Errorf("project %s not found in BG", poPrj.ProjectCode)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE projects SET name=$1, division=$2, department=$3, project_group=$4, item_no=$5
		WHERE id=$6`,
		poPrj.Name, poPrj.Division, poPrj.Department, poPrj.GroupName, poPrj.ItemNo, projectID); err != nil {
		return err
	}

	for _, sj := range poPrj.SubJobs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO sub_jobs (project_id, name, fund_type, data_year, budget, target, cut_transfer, under_budget)
			VALUES ($1,$2,$3,$4,$5,$6,0,0)
			ON CONFLICT (project_id, name, fund_type, data_year) DO UPDATE
			SET budget=$5, target=$6`,
			projectID, sj.Name, sj.FundType, sj.DataYear, sj.Budget, sj.Target); err != nil {
			return err
		}
	}

	snapshot, _ := json.Marshal(poPrj)
	if _, err := tx.Exec(ctx, `
		INSERT INTO po_import_log (project_code, po_version, accepted_by, snapshot_json)
		VALUES ($1,$2,'system',$3)`,
		poPrj.ProjectCode, poPrj.Version, snapshot); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO po_import_status (project_code, last_accepted_version, last_accepted_at, po_version, po_updated_at, status)
		VALUES ($1,$2,$3,$2,$3,'up_to_date')
		ON CONFLICT (project_code) DO UPDATE
		SET last_accepted_version=$2, last_accepted_at=$3, status='up_to_date'`,
		poPrj.ProjectCode, poPrj.Version, time.Now()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func diffProjectFields(bg models.Project, p models.POProject) []models.FieldDiff {
	var diffs []models.FieldDiff
	if bg.Name != p.Name {
		diffs = append(diffs, models.FieldDiff{Field: "name", BGValue: bg.Name, POValue: p.Name})
	}
	if ptrStr(bg.Division) != ptrStr(p.Division) {
		diffs = append(diffs, models.FieldDiff{Field: "division", BGValue: ptrStr(bg.Division), POValue: ptrStr(p.Division)})
	}
	if ptrStr(bg.Department) != ptrStr(p.Department) {
		diffs = append(diffs, models.FieldDiff{Field: "department", BGValue: ptrStr(bg.Department), POValue: ptrStr(p.Department)})
	}
	if ptrStr(bg.GroupName) != ptrStr(p.GroupName) {
		diffs = append(diffs, models.FieldDiff{Field: "group_name", BGValue: ptrStr(bg.GroupName), POValue: ptrStr(p.GroupName)})
	}
	if ptrStr(bg.ItemNo) != ptrStr(p.ItemNo) {
		diffs = append(diffs, models.FieldDiff{Field: "item_no", BGValue: ptrStr(bg.ItemNo), POValue: ptrStr(p.ItemNo)})
	}
	return diffs
}

func diffSubJobFields(bg, p models.POSubJob) []models.FieldDiff {
	var diffs []models.FieldDiff
	if bg.Budget != p.Budget {
		diffs = append(diffs, models.FieldDiff{Field: "budget", BGValue: bg.Budget, POValue: p.Budget})
	}
	if bg.Target != p.Target {
		diffs = append(diffs, models.FieldDiff{Field: "target", BGValue: bg.Target, POValue: p.Target})
	}
	return diffs
}

func subJobDiffsAllAdded(sjs []models.POSubJob) []models.SubJobDiff {
	diffs := make([]models.SubJobDiff, len(sjs))
	for i, sj := range sjs {
		diffs[i] = models.SubJobDiff{
			Name: sj.Name, FundType: sj.FundType, DataYear: sj.DataYear,
			Change: "added",
			Diffs:  []models.FieldDiff{{Field: "budget", BGValue: nil, POValue: sj.Budget}, {Field: "target", BGValue: nil, POValue: sj.Target}},
		}
	}
	return diffs
}

func hasChangedSubJobs(diffs []models.SubJobDiff) bool {
	for _, d := range diffs {
		if d.Change != "unchanged" {
			return true
		}
	}
	return false
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

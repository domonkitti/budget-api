package handlers

import (
	"net/http"
	"strconv"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SummaryHandler struct {
	db *pgxpool.Pool
}

func NewSummaryHandler(db *pgxpool.Pool) *SummaryHandler {
	return &SummaryHandler{db: db}
}

// GET /summary?by=division&year=2570&fund_type=ลงทุน&source=เงินกู้
func (h *SummaryHandler) Summarize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	by := q.Get("by") // division | project_type | source
	year := q.Get("year")
	fundType := q.Get("fund_type")
	source := q.Get("source")

	groupCol := map[string]string{
		"division":     "p.division",
		"project_type": "p.project_type",
		"source":       "bs.source",
	}[by]

	if groupCol == "" {
		http.Error(w, "by must be division, project_type, or source", http.StatusBadRequest)
		return
	}

	sql := `
		SELECT ` + groupCol + ` as group_by,
		       SUM(bs.budget) as budget,
		       SUM(bs.target) as target,
		       SUM(bs.budget - bs.target) as remain
		FROM budget_sources bs
		JOIN projects p ON p.id = bs.project_id
		WHERE 1=1`
	args := []any{}
	i := 1

	if year != "" {
		sql += ` AND p.year = $` + itoa(i)
		args = append(args, year)
		i++
	}
	if fundType != "" {
		sql += ` AND bs.fund_type = $` + itoa(i)
		args = append(args, fundType)
		i++
	}
	if source != "" {
		sql += ` AND bs.source = $` + itoa(i)
		args = append(args, source)
		i++
	}

	sql += ` GROUP BY ` + groupCol + ` ORDER BY budget DESC`

	rows, err := h.db.Query(r.Context(), sql, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := []models.SummaryRow{}
	for rows.Next() {
		var row models.SummaryRow
		rows.Scan(&row.GroupBy, &row.Budget, &row.Target, &row.Remain)
		result = append(result, row)
	}

	respond(w, http.StatusOK, result)
}

// GET /summary/top?by=name&limit=10&source=เงินกู้&fund_type=ลงทุน
func (h *SummaryHandler) TopN(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := q.Get("limit")
	fundType := q.Get("fund_type")
	source := q.Get("source")
	year := q.Get("year")

	if limit == "" {
		limit = "10"
	}

	sql := `
		SELECT p.name as group_by,
		       SUM(bs.budget) as budget,
		       SUM(bs.target) as target,
		       SUM(bs.budget - bs.target) as remain
		FROM budget_sources bs
		JOIN projects p ON p.id = bs.project_id
		WHERE 1=1`
	args := []any{}
	i := 1

	if year != "" {
		sql += ` AND p.year = $` + itoa(i)
		args = append(args, year)
		i++
	}
	if fundType != "" {
		sql += ` AND bs.fund_type = $` + itoa(i)
		args = append(args, fundType)
		i++
	}
	if source != "" {
		sql += ` AND bs.source = $` + itoa(i)
		args = append(args, source)
		i++
	}

	sql += ` GROUP BY p.name ORDER BY target DESC LIMIT $` + itoa(i)
	args = append(args, limit)

	rows, err := h.db.Query(r.Context(), sql, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := []models.SummaryRow{}
	for rows.Next() {
		var row models.SummaryRow
		rows.Scan(&row.GroupBy, &row.Budget, &row.Target, &row.Remain)
		result = append(result, row)
	}

	respond(w, http.StatusOK, result)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}


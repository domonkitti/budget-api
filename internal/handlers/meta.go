package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MetaHandler struct {
	db *pgxpool.Pool
}

func NewMetaHandler(db *pgxpool.Pool) *MetaHandler {
	return &MetaHandler{db: db}
}

type filterOptions struct {
	Years       []int    `json:"years"`
	Sources     []string `json:"sources"`
	Divisions   []string `json:"divisions"`
	Departments []string `json:"departments"`
	Groups      []string `json:"groups"`
}

func (h *MetaHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	scanStrings := func(query string) []string {
		out := []string{}
		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var s string
			rows.Scan(&s)
			out = append(out, s)
		}
		return out
	}

	years := []int{}
	rows, err := h.db.Query(r.Context(),
		`SELECT DISTINCT data_year FROM sub_jobs ORDER BY data_year`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var y int
		rows.Scan(&y)
		years = append(years, y)
	}

	respond(w, http.StatusOK, filterOptions{
		Years:       years,
		Sources:     scanStrings(`SELECT DISTINCT source FROM budget_sources WHERE source IS NOT NULL ORDER BY source`),
		Divisions:   scanStrings(`SELECT DISTINCT division FROM projects WHERE division IS NOT NULL ORDER BY division`),
		Departments: scanStrings(`SELECT DISTINCT department FROM projects WHERE department IS NOT NULL ORDER BY department`),
		Groups:      scanStrings(`SELECT DISTINCT project_group FROM projects WHERE project_group IS NOT NULL ORDER BY project_group`),
	})
}

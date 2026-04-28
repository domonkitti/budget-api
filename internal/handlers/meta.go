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
	Years   []int    `json:"years"`
	Sources []string `json:"sources"`
}

func (h *MetaHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
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

	sources := []string{}
	srows, err := h.db.Query(r.Context(),
		`SELECT DISTINCT source FROM budget_sources ORDER BY source`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer srows.Close()
	for srows.Next() {
		var s string
		srows.Scan(&s)
		sources = append(sources, s)
	}

	respond(w, http.StatusOK, filterOptions{Years: years, Sources: sources})
}

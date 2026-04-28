package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
)

// mockStore holds all data loaded from mock_data.json at startup.
type mockStore struct {
	Projects      []mockProject      `json:"projects"`
	SubJobs       []mockSubJob       `json:"sub_jobs"`
	BudgetSources []mockBudgetSource `json:"budget_sources"`
	TagCategories []mockTagCategory  `json:"tag_categories"`
	TagValues     []mockTagValue     `json:"tag_values"`
}

type mockProject struct {
	ID          int     `json:"id"`
	ProjectCode string  `json:"project_code"`
	Year        int     `json:"year"`
	ProjectType string  `json:"project_type"`
	ItemNo      *string `json:"item_no"`
	Name        string  `json:"name"`
	Division    *string `json:"division"`
	Department  *string `json:"department"`
}

type mockSubJob struct {
	ID        int     `json:"id"`
	ProjectID int     `json:"project_id"`
	Name      string  `json:"name"`
	SortOrder *int    `json:"sort_order"`
	FundType  string  `json:"fund_type"`
	DataYear  int     `json:"data_year"`
	Budget    float64 `json:"budget"`
	Target    float64 `json:"target"`
}

type mockBudgetSource struct {
	ID        int     `json:"id"`
	ProjectID int     `json:"project_id"`
	Source    string  `json:"source"`
	FundType  string  `json:"fund_type"`
	DataYear  int     `json:"data_year"`
	Budget    float64 `json:"budget"`
	Target    float64 `json:"target"`
}

type mockTagCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type mockTagValue struct {
	ID         int    `json:"id"`
	CategoryID int    `json:"category_id"`
	Code       string `json:"code"`
}

func LoadMockStore(path string) *mockStore {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("mock: read %s: %v", path, err)
	}
	var s mockStore
	if err := json.Unmarshal(data, &s); err != nil {
		log.Fatalf("mock: parse %s: %v", path, err)
	}
	log.Printf("mock: loaded %d projects from %s", len(s.Projects), path)
	return &s
}

var mockTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ── Mock Project Handler ───────────────────────────────────────────────────────

type MockProjectHandler struct{ s *mockStore }

func NewMockProjectHandler(s *mockStore) *MockProjectHandler { return &MockProjectHandler{s: s} }

func (h *MockProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, typ, div := q.Get("year"), q.Get("type"), q.Get("division")

	result := []models.Project{}
	for _, p := range h.s.Projects {
		if year != "" && itoa(p.Year) != year {
			continue
		}
		if typ != "" && p.ProjectType != typ {
			continue
		}
		if div != "" && (p.Division == nil || *p.Division != div) {
			continue
		}
		result = append(result, toProject(p))
	}
	respond(w, http.StatusOK, result)
}

func (h *MockProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	for _, p := range h.s.Projects {
		if p.ProjectCode != code {
			continue
		}
		detail := models.ProjectDetail{
			Project:       toProject(p),
			SubJobs:       []models.SubJob{},
			BudgetSources: []models.BudgetSource{},
		}
		for _, sj := range h.s.SubJobs {
			if sj.ProjectID == p.ID {
				detail.SubJobs = append(detail.SubJobs, toSubJob(sj))
			}
		}
		for _, bs := range h.s.BudgetSources {
			if bs.ProjectID == p.ID {
				detail.BudgetSources = append(detail.BudgetSources, toBudgetSource(bs))
			}
		}
		respond(w, http.StatusOK, detail)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (h *MockProjectHandler) Flat(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, typ, div, source := q.Get("year"), q.Get("type"), q.Get("division"), q.Get("source")

	sourceProjectIDs := map[int]bool{}
	if source != "" {
		for _, bs := range h.s.BudgetSources {
			if bs.Source == source {
				sourceProjectIDs[bs.ProjectID] = true
			}
		}
	}

	result := []models.FlatProject{}
	for _, p := range h.s.Projects {
		if year != "" && itoa(p.Year) != year {
			continue
		}
		if typ != "" && p.ProjectType != typ {
			continue
		}
		if div != "" && (p.Division == nil || *p.Division != div) {
			continue
		}
		if source != "" && !sourceProjectIDs[p.ID] {
			continue
		}

		// Build source_breakdown: aggregate by (year, source, fund type)
		type key struct {
			year             int
			source, fundType string
		}
		agg := map[key]*models.SourceYearEntry{}
		for _, bs := range h.s.BudgetSources {
			if bs.ProjectID != p.ID {
				continue
			}
			k := key{bs.DataYear, bs.Source, bs.FundType}
			if agg[k] == nil {
				agg[k] = &models.SourceYearEntry{Year: bs.DataYear, Source: bs.Source, FundType: bs.FundType}
			}
			agg[k].Budget += bs.Budget
			agg[k].Target += bs.Target
			agg[k].Remain += bs.Budget - bs.Target
		}
		breakdown := []models.SourceYearEntry{}
		for _, e := range agg {
			breakdown = append(breakdown, *e)
		}

		type subJobKey struct {
			name      string
			sortOrder *int
			year      int
			fundType  string
		}
		subJobAgg := map[subJobKey]*models.SubJobYearEntry{}
		for _, sj := range h.s.SubJobs {
			if sj.ProjectID != p.ID {
				continue
			}
			k := subJobKey{sj.Name, sj.SortOrder, sj.DataYear, sj.FundType}
			if subJobAgg[k] == nil {
				subJobAgg[k] = &models.SubJobYearEntry{
					Name: sj.Name, SortOrder: sj.SortOrder, Year: sj.DataYear, FundType: sj.FundType,
				}
			}
			subJobAgg[k].Budget += sj.Budget
			subJobAgg[k].Target += sj.Target
			subJobAgg[k].Remain += sj.Budget - sj.Target
		}
		subJobs := []models.SubJobYearEntry{}
		for _, e := range subJobAgg {
			subJobs = append(subJobs, *e)
		}

		result = append(result, models.FlatProject{
			ID: p.ID, ProjectCode: p.ProjectCode, ItemNo: p.ItemNo, Name: p.Name,
			Division: p.Division, ProjectType: p.ProjectType, Year: p.Year,
			SubJobs: subJobs, SourceBreakdown: breakdown,
		})
	}
	respond(w, http.StatusOK, result)
}

// ── Mock Summary Handler ───────────────────────────────────────────────────────

type MockSummaryHandler struct{ s *mockStore }

func NewMockSummaryHandler(s *mockStore) *MockSummaryHandler { return &MockSummaryHandler{s: s} }

func (h *MockSummaryHandler) Summarize(w http.ResponseWriter, r *http.Request) {
	by := r.URL.Query().Get("by")
	if by != "division" && by != "project_type" && by != "source" {
		http.Error(w, "by must be division, project_type, or source", http.StatusBadRequest)
		return
	}

	totals := map[string]*models.SummaryRow{}
	projectByID := map[int]mockProject{}
	for _, p := range h.s.Projects {
		projectByID[p.ID] = p
	}

	for _, bs := range h.s.BudgetSources {
		p := projectByID[bs.ProjectID]
		var key string
		switch by {
		case "division":
			if p.Division != nil {
				key = *p.Division
			}
		case "project_type":
			key = p.ProjectType
		case "source":
			key = bs.Source
		}
		if key == "" {
			continue
		}
		if totals[key] == nil {
			totals[key] = &models.SummaryRow{GroupBy: key}
		}
		totals[key].Budget += bs.Budget
		totals[key].Target += bs.Target
		totals[key].Remain += bs.Budget - bs.Target
	}

	result := []models.SummaryRow{}
	for _, row := range totals {
		result = append(result, *row)
	}
	respond(w, http.StatusOK, result)
}

func (h *MockSummaryHandler) TopN(w http.ResponseWriter, r *http.Request) {
	projectByID := map[int]mockProject{}
	for _, p := range h.s.Projects {
		projectByID[p.ID] = p
	}
	totals := map[int]*models.SummaryRow{}
	for _, bs := range h.s.BudgetSources {
		if totals[bs.ProjectID] == nil {
			totals[bs.ProjectID] = &models.SummaryRow{GroupBy: projectByID[bs.ProjectID].Name}
		}
		totals[bs.ProjectID].Budget += bs.Budget
		totals[bs.ProjectID].Target += bs.Target
		totals[bs.ProjectID].Remain += bs.Budget - bs.Target
	}
	result := []models.SummaryRow{}
	for _, row := range totals {
		result = append(result, *row)
	}
	respond(w, http.StatusOK, result)
}

// ── Mock Tag Handler ───────────────────────────────────────────────────────────

type MockTagHandler struct{ s *mockStore }

func NewMockTagHandler(s *mockStore) *MockTagHandler { return &MockTagHandler{s: s} }

func (h *MockTagHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats := make([]models.TagCategory, len(h.s.TagCategories))
	for i, c := range h.s.TagCategories {
		cats[i] = models.TagCategory{ID: c.ID, Name: c.Name, CreatedAt: mockTime}
	}
	respond(w, http.StatusOK, cats)
}

func (h *MockTagHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusCreated, models.TagCategory{ID: 99, Name: "mock", CreatedAt: mockTime})
}

func (h *MockTagHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *MockTagHandler) ListValues(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catID")
	vals := []models.TagValue{}
	for _, v := range h.s.TagValues {
		if itoa(v.CategoryID) == catID {
			vals = append(vals, models.TagValue{ID: v.ID, CategoryID: v.CategoryID, Code: v.Code, CreatedAt: mockTime})
		}
	}
	respond(w, http.StatusOK, vals)
}

func (h *MockTagHandler) CreateValue(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusCreated, models.TagValue{ID: 99, CategoryID: 1, Code: "mock", CreatedAt: mockTime})
}

func (h *MockTagHandler) UpdateValue(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, models.TagValue{ID: 99, CategoryID: 1, Code: "mock", CreatedAt: mockTime})
}

func (h *MockTagHandler) DeleteValue(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *MockTagHandler) GetProjectTags(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, []models.ProjectTag{})
}

func (h *MockTagHandler) SetProjectTags(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *MockTagHandler) GetSubJobTags(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, []models.SubJobTag{})
}

func (h *MockTagHandler) SetSubJobTags(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *MockTagHandler) GetAllocationSelections(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, []models.CategoryAllocationSelection{})
}

func (h *MockTagHandler) SetAllocationSelections(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *MockTagHandler) SummaryByTag(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, []models.TagSummaryRow{
		{Code: "SO1", Budget: 5000000, Target: 3000000, Remain: 2000000},
		{Code: "SO2", Budget: 3000000, Target: 1500000, Remain: 1500000},
	})
}

// ── Mock Meta Handler ─────────────────────────────────────────────────────────

type MockMetaHandler struct{ s *mockStore }

func NewMockMetaHandler(s *mockStore) *MockMetaHandler { return &MockMetaHandler{s: s} }

func (h *MockMetaHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	yearSet := map[int]struct{}{}
	for _, sj := range h.s.SubJobs {
		yearSet[sj.DataYear] = struct{}{}
	}
	years := []int{}
	for y := range yearSet {
		years = append(years, y)
	}
	// simple sort
	for i := 0; i < len(years)-1; i++ {
		for j := i + 1; j < len(years); j++ {
			if years[i] > years[j] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	sourceSet := map[string]struct{}{}
	for _, bs := range h.s.BudgetSources {
		sourceSet[bs.Source] = struct{}{}
	}
	sources := []string{}
	for s := range sourceSet {
		sources = append(sources, s)
	}

	respond(w, http.StatusOK, filterOptions{Years: years, Sources: sources})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toProject(p mockProject) models.Project {
	return models.Project{
		ID: p.ID, ProjectCode: p.ProjectCode, Year: p.Year,
		ProjectType: p.ProjectType, ItemNo: p.ItemNo, Name: p.Name,
		Division: p.Division, Department: p.Department, CreatedAt: mockTime,
	}
}

func toSubJob(sj mockSubJob) models.SubJob {
	return models.SubJob{
		ID: sj.ID, ProjectID: sj.ProjectID, Name: sj.Name,
		SortOrder: sj.SortOrder, FundType: sj.FundType, DataYear: sj.DataYear,
		Budget: sj.Budget, Target: sj.Target, Remain: sj.Budget - sj.Target,
		CreatedAt: mockTime,
	}
}

func toBudgetSource(bs mockBudgetSource) models.BudgetSource {
	return models.BudgetSource{
		ID: bs.ID, ProjectID: bs.ProjectID, Source: bs.Source,
		FundType: bs.FundType, DataYear: bs.DataYear,
		Budget: bs.Budget, Target: bs.Target,
		Remain: bs.Budget - bs.Target, CreatedAt: mockTime,
	}
}

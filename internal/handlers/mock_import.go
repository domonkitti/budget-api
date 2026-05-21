package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
)

// mockProjectOverview is used by ProjectOverview in MOCK mode.
var mockProjectOverview = []models.ProjectOverviewItem{
	// C type, year 2568 < active 2570 → แผนงานต่อเนื่อง; has sub_jobs in BG
	{
		ProjectCode: "I2568C001", Name: "แผนงานสนับสนุนการดำเนินงาน ระยะที่ 6",
		ProjectType: "C", ProjectYear: 2568, Status: "has_update",
		FullPlanBudget:   1715.199,
		ActiveYearBudget: 1715.040,
	},
	// C type, year 2570 = active → แผนงานใหม่; BG has budget_sources only, PO has sub_jobs
	{
		ProjectCode: "I2570C003", Name: "แผนงานการก่อสร้างปรับปรุงระบบจำหน่ายไฟฟ้าเป็นเคเบิลใต้ดิน",
		ProjectType: "C", ProjectYear: 2570, Status: "new",
		FullPlanBudget:   850.000,
		ActiveYearBudget: 850.000,
	},
	// C type, year 2570 = active → แผนงานใหม่; BG has budget_sources only, PO also has no sub_jobs
	{
		ProjectCode: "I2570C004", Name: "แผนงานเช่ายานพาหนะ ปี 2570 - 2575",
		ProjectType: "C", ProjectYear: 2570, Status: "up_to_date",
		FullPlanBudget:   3919.947,
		ActiveYearBudget: 1306.649,
	},
	// C type, year 2570 = active → แผนงานใหม่; no BG record at all, budgets from PO
	{
		ProjectCode: "I2570C005", Name: "แผนงานพัฒนาระบบพลังงานทดแทน ระยะที่ 1",
		ProjectType: "C", ProjectYear: 2570, Status: "new",
		FullPlanBudget:   220000.000, // PO: 15000 + 85000 + 120000
		ActiveYearBudget: 100000.000, // PO 2570 only: 15000 + 85000
	},
}

type MockImportHandler struct{}

func NewMockImportHandler() *MockImportHandler { return &MockImportHandler{} }

var (
	mockImportStatuses = func() []models.ImportStatus {
		v1 := 1
		v2 := 2
		accV1 := 1
		tC001  := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
		tC003  := time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)
		tC004  := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
		tC005  := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
		accT   := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		return []models.ImportStatus{
			// I2568C001: BG has accepted v1, PO now at v2 → has_update
			{ProjectCode: "I2568C001", Status: "has_update",
				LastAcceptedVersion: &accV1, LastAcceptedAt: &accT,
				POVersion: &v2, POUpdatedAt: &tC001},
			// I2570C003: BG has budget_sources, PO has sub_jobs, never accepted → new
			{ProjectCode: "I2570C003", Status: "new",
				LastAcceptedVersion: nil, LastAcceptedAt: nil,
				POVersion: &v1, POUpdatedAt: &tC003},
			// I2570C004: BG has budget_sources, PO also has no sub_jobs, already synced → up_to_date
			{ProjectCode: "I2570C004", Status: "up_to_date",
				LastAcceptedVersion: &v1, LastAcceptedAt: &accT,
				POVersion: &v1, POUpdatedAt: &tC004},
			// I2570C005: no BG record at all → new
			{ProjectCode: "I2570C005", Status: "new",
				LastAcceptedVersion: nil, LastAcceptedAt: nil,
				POVersion: &v1, POUpdatedAt: &tC005},
		}
	}()

	mockDiffs = map[string]models.ProjectDiff{
		// I2568C001: two sub_job budgets changed in ลงทุน rows
		"I2568C001": {
			ProjectCode:  "I2568C001",
			POVersion:    2,
			HasChanges:   true,
			ProjectDiffs: []models.FieldDiff{},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "1. หมวดที่ดิน",               FundType: "ลงทุน",  DataYear: 2570, Change: "unchanged"},
				{Name: "2. หมวดสิ่งก่อสร้าง",         FundType: "ผูกพัน", DataYear: 2570, Change: "unchanged"},
				{Name: "2. หมวดสิ่งก่อสร้าง",         FundType: "ลงทุน",  DataYear: 2570, Change: "modified",
					Diffs: []models.FieldDiff{{Field: "budget", BGValue: 414.884, POValue: 520.000}}},
				{Name: "3. หมวดยานพาหนะ",             FundType: "ผูกพัน", DataYear: 2570, Change: "unchanged"},
				{Name: "3. หมวดยานพาหนะ",             FundType: "ลงทุน",  DataYear: 2570, Change: "modified",
					Diffs: []models.FieldDiff{{Field: "budget", BGValue: 799.275, POValue: 850.000}}},
				{Name: "4. หมวดเครื่องมือ-เครื่องใช้", FundType: "ผูกพัน", DataYear: 2570, Change: "unchanged"},
				{Name: "4. หมวดเครื่องมือ-เครื่องใช้", FundType: "ลงทุน",  DataYear: 2570, Change: "unchanged"},
				{Name: "สำรองราคา",                   FundType: "ผูกพัน", DataYear: 2570, Change: "unchanged"},
			},
		},
		// I2570C003: BG has no sub_jobs → PO sub_jobs all appear as "added"
		"I2570C003": {
			ProjectCode:  "I2570C003",
			POVersion:    1,
			HasChanges:   true,
			ProjectDiffs: []models.FieldDiff{},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "งานรวม", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 850.000},
						{Field: "target", BGValue: nil, POValue: 78.45},
					}},
				{Name: "งานรวม", FundType: "ลงทุน", DataYear: 2571, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 920.000},
						{Field: "target", BGValue: nil, POValue: 0.0},
					}},
			},
		},
		// I2570C004: both BG and PO have no sub_jobs → no changes
		"I2570C004": {
			ProjectCode:  "I2570C004",
			POVersion:    1,
			HasChanges:   false,
			ProjectDiffs: []models.FieldDiff{},
			SubJobDiffs:  []models.SubJobDiff{},
		},
		// I2570C005: no BG record at all → project itself is "added"
		"I2570C005": {
			ProjectCode: "I2570C005",
			POVersion:   1,
			HasChanges:  true,
			ProjectDiffs: []models.FieldDiff{
				{Field: "project", BGValue: nil, POValue: "new project"},
			},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "1. งานสำรวจและออกแบบ", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 15000.000},
						{Field: "target", BGValue: nil, POValue: 0.0},
					}},
				{Name: "2. งานก่อสร้าง", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 85000.000},
						{Field: "target", BGValue: nil, POValue: 0.0},
					}},
				{Name: "2. งานก่อสร้าง", FundType: "ลงทุน", DataYear: 2571, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 120000.000},
						{Field: "target", BGValue: nil, POValue: 0.0},
					}},
			},
		},
	}

	mockImportLogs = []models.ImportLog{
		{ID: 1, ProjectCode: "I2568C001", POVersion: 1, AcceptedBy: "system",
			AcceptedAt: time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC), SnapshotJSON: []byte(`{}`)},
		{ID: 2, ProjectCode: "I2570C004", POVersion: 1, AcceptedBy: "system",
			AcceptedAt: time.Date(2026, 3, 1, 9, 5, 0, 0, time.UTC), SnapshotJSON: []byte(`{}`)},
	}
)

func (h *MockImportHandler) CheckVersions(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, mockImportStatuses)
}

func (h *MockImportHandler) FetchDiff(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	diff, ok := mockDiffs[code]
	if !ok {
		http.Error(w, "project not found in PO system", http.StatusNotFound)
		return
	}
	respond(w, http.StatusOK, diff)
}

func (h *MockImportHandler) Accept(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	respond(w, http.StatusOK, map[string]any{"ok": true, "project_code": code, "po_version": 2})
}

func (h *MockImportHandler) BatchAccept(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]any{"results": []map[string]any{
		{"project_code": "I2568C001", "ok": true, "po_version": 2},
	}})
}

func (h *MockImportHandler) ListLog(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, mockImportLogs)
}

func (h *MockImportHandler) ProjectOverview(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	year := 2570
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	result := []models.ProjectOverviewItem{}
	for _, item := range mockProjectOverview {
		if item.ActiveYearBudget > 0 || year == 2570 {
			result = append(result, item)
		}
	}
	respond(w, http.StatusOK, result)
}

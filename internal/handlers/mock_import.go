package handlers

import (
	"net/http"
	"time"

	"github.com/domonkitti/budget-app-api/internal/models"
	"github.com/go-chi/chi/v5"
)

type MockImportHandler struct{}

func NewMockImportHandler() *MockImportHandler { return &MockImportHandler{} }

var (
	mockImportStatuses = func() []models.ImportStatus {
		v2, v3 := 2, 3
		v1 := 1
		t1 := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 4, 10, 14, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 5, 15, 11, 0, 0, 0, time.UTC)
		accV := 2
		accT := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		return []models.ImportStatus{
			{ProjectCode: "I2570Y001", Status: "has_update", LastAcceptedVersion: &accV, LastAcceptedAt: &accT, POVersion: &v3, POUpdatedAt: &t1},
			{ProjectCode: "I2570Y002", Status: "up_to_date", LastAcceptedVersion: &accV, LastAcceptedAt: &accT, POVersion: &v2, POUpdatedAt: &t2},
			{ProjectCode: "I2570C001", Status: "new", LastAcceptedVersion: nil, LastAcceptedAt: nil, POVersion: &v1, POUpdatedAt: &t3},
		}
	}()

	mockDiffs = map[string]models.ProjectDiff{
		"I2570Y001": {
			ProjectCode: "I2570Y001",
			POVersion:   3,
			HasChanges:  true,
			ProjectDiffs: []models.FieldDiff{
				{Field: "name", BGValue: "โครงการปรับปรุงระบบน้ำประปา", POValue: "โครงการปรับปรุงระบบน้ำประปา (ฉบับปรับปรุง)"},
			},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "งานออกแบบ", FundType: "ลงทุน", DataYear: 2570, Change: "unchanged"},
				{Name: "งานก่อสร้าง", FundType: "ลงทุน", DataYear: 2570, Change: "modified",
					Diffs: []models.FieldDiff{{Field: "budget", BGValue: 7000000.0, POValue: 8000000.0}}},
				{Name: "งานก่อสร้าง", FundType: "ผูกพัน", DataYear: 2570, Change: "unchanged"},
				{Name: "งานตรวจสอบ", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{
						{Field: "budget", BGValue: nil, POValue: 500000.0},
						{Field: "target", BGValue: nil, POValue: 400000.0},
					}},
			},
		},
		"I2570Y002": {
			ProjectCode:  "I2570Y002",
			POVersion:    2,
			HasChanges:   false,
			ProjectDiffs: []models.FieldDiff{},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "งานฐานราก", FundType: "ลงทุน", DataYear: 2570, Change: "unchanged"},
				{Name: "งานโครงสร้าง", FundType: "ลงทุน", DataYear: 2570, Change: "unchanged"},
			},
		},
		"I2570C001": {
			ProjectCode: "I2570C001",
			POVersion:   1,
			HasChanges:  true,
			ProjectDiffs: []models.FieldDiff{
				{Field: "project", BGValue: nil, POValue: "new project"},
			},
			SubJobDiffs: []models.SubJobDiff{
				{Name: "งานวิเคราะห์ความต้องการ", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{{Field: "budget", BGValue: nil, POValue: 800000.0}, {Field: "target", BGValue: nil, POValue: 700000.0}}},
				{Name: "งานพัฒนาระบบ", FundType: "ลงทุน", DataYear: 2570, Change: "added",
					Diffs: []models.FieldDiff{{Field: "budget", BGValue: nil, POValue: 4500000.0}, {Field: "target", BGValue: nil, POValue: 4000000.0}}},
			},
		},
	}

	mockImportLogs = []models.ImportLog{
		{ID: 1, ProjectCode: "I2570Y001", POVersion: 2, AcceptedBy: "system",
			AcceptedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC), SnapshotJSON: []byte(`{}`)},
		{ID: 2, ProjectCode: "I2570Y002", POVersion: 2, AcceptedBy: "system",
			AcceptedAt: time.Date(2026, 4, 1, 10, 5, 0, 0, time.UTC), SnapshotJSON: []byte(`{}`)},
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
	respond(w, http.StatusOK, map[string]any{"ok": true, "project_code": code, "po_version": 3})
}

func (h *MockImportHandler) BatchAccept(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]any{"results": []map[string]any{
		{"project_code": "I2570Y001", "ok": true, "po_version": 3},
	}})
}

func (h *MockImportHandler) ListLog(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, mockImportLogs)
}

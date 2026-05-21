package models

import (
	"encoding/json"
	"time"
)

// POProjectVersion is one entry from the PO versions endpoint.
type POProjectVersion struct {
	ProjectCode string    `json:"project_code"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// POSubJob is a sub-job row as sent by the PO system.
type POSubJob struct {
	Name     string  `json:"name"`
	FundType string  `json:"fund_type"`
	DataYear int     `json:"data_year"`
	Budget   float64 `json:"budget"`
	Target   float64 `json:"target"`
}

// POProject is a full project export from the PO system.
type POProject struct {
	ProjectCode string     `json:"project_code"`
	Version     int        `json:"version"`
	ExportedAt  time.Time  `json:"exported_at"`
	Name        string     `json:"name"`
	Division    *string    `json:"division"`
	Department  *string    `json:"department"`
	GroupName   *string    `json:"group_name"`
	ItemNo      *string    `json:"item_no"`
	SubJobs     []POSubJob `json:"sub_jobs"`
}

// ImportStatus is the current sync state of one project with the PO system.
type ImportStatus struct {
	ProjectCode         string     `json:"project_code"`
	Status              string     `json:"status"` // "up_to_date" | "has_update" | "new"
	LastAcceptedVersion *int       `json:"last_accepted_version"`
	LastAcceptedAt      *time.Time `json:"last_accepted_at"`
	POVersion           *int       `json:"po_version"`
	POUpdatedAt         *time.Time `json:"po_updated_at"`
}

// ImportLog is one accepted-import audit entry.
type ImportLog struct {
	ID           int             `json:"id"`
	ProjectCode  string          `json:"project_code"`
	POVersion    int             `json:"po_version"`
	AcceptedBy   string          `json:"accepted_by"`
	AcceptedAt   time.Time       `json:"accepted_at"`
	SnapshotJSON json.RawMessage `json:"snapshot_json"`
}

// FieldDiff describes one changed project-level field.
type FieldDiff struct {
	Field   string `json:"field"`
	BGValue any    `json:"bg_value"`
	POValue any    `json:"po_value"`
}

// SubJobDiff describes the diff for one sub-job row.
type SubJobDiff struct {
	Name     string      `json:"name"`
	FundType string      `json:"fund_type"`
	DataYear int         `json:"data_year"`
	Change   string      `json:"change"` // "added" | "modified" | "removed" | "unchanged"
	Diffs    []FieldDiff `json:"diffs,omitempty"`
}

// ProjectDiff is the full diff result for one project.
type ProjectDiff struct {
	ProjectCode  string       `json:"project_code"`
	POVersion    int          `json:"po_version"`
	HasChanges   bool         `json:"has_changes"`
	ProjectDiffs []FieldDiff  `json:"project_diffs"`
	SubJobDiffs  []SubJobDiff `json:"sub_job_diffs"`
}

// ProjectOverviewItem is one row in the project overview page.
type ProjectOverviewItem struct {
	ProjectCode      string  `json:"project_code"`
	Name             string  `json:"name"`
	ProjectType      string  `json:"project_type"`    // Y | C | L
	ProjectYear      int     `json:"project_year"`     // year from project code (start year)
	GroupName        *string `json:"group_name"`       // หมวด (Y type only)
	ItemNo           *string `json:"item_no"`
	Status           string  `json:"status"`           // "has_update" | "new" | "up_to_date" | "budget_only"
	FullPlanBudget   float64 `json:"full_plan_budget"`
	ActiveYearBudget float64 `json:"active_year_budget"`
}

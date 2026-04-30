package models

import "time"

type Project struct {
	ID          int       `json:"id"`
	ProjectCode string    `json:"project_code"`
	Year        int       `json:"year"`
	ProjectType string    `json:"project_type"`
	ItemNo      *string   `json:"item_no,omitempty"`
	Name        string    `json:"name"`
	Division    *string   `json:"division,omitempty"`
	Department  *string   `json:"department,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type SubJob struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Name      string    `json:"name"`
	SortOrder *int      `json:"sort_order,omitempty"`
	FundType  string    `json:"fund_type"`
	DataYear  int       `json:"data_year"`
	Budget    float64   `json:"budget"`
	Target    float64   `json:"target"`
	Remain    float64   `json:"remain"`
	CreatedAt time.Time `json:"created_at"`
}

type BudgetSource struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Source    string    `json:"source"`
	FundType  string    `json:"fund_type"`
	DataYear  int       `json:"data_year"`
	Budget    float64   `json:"budget"`
	Target    float64   `json:"target"`
	Remain    float64   `json:"remain"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectDetail struct {
	Project
	SubJobs       []SubJob       `json:"sub_jobs"`
	BudgetSources []BudgetSource `json:"budget_sources"`
}

type SummaryRow struct {
	GroupBy string  `json:"group_by"`
	Budget  float64 `json:"budget"`
	Target  float64 `json:"target"`
	Remain  float64 `json:"remain"`
}

type TagCategory struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TagValue struct {
	ID         int       `json:"id"`
	CategoryID int       `json:"category_id"`
	Code       string    `json:"code"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubJobTag struct {
	ID         int     `json:"id"`
	ProjectID  int     `json:"project_id"`
	SubJobName string  `json:"sub_job_name"`
	TagValueID int     `json:"tag_value_id"`
	TagCode    string  `json:"tag_code,omitempty"`
	CategoryID int     `json:"category_id,omitempty"`
	Percentage float64 `json:"percentage"`
}

type ProjectTag struct {
	ID         int     `json:"id"`
	ProjectID  int     `json:"project_id"`
	TagValueID int     `json:"tag_value_id"`
	TagCode    string  `json:"tag_code,omitempty"`
	CategoryID int     `json:"category_id,omitempty"`
	Percentage float64 `json:"percentage"`
}

type CategoryAllocationSelection struct {
	ID         int     `json:"id,omitempty"`
	CategoryID int     `json:"category_id"`
	ProjectID  int     `json:"project_id"`
	TargetType string  `json:"target_type"`
	SubJobName *string `json:"sub_job_name,omitempty"`
}

type TagSummaryRow struct {
	Code   string  `json:"code"`
	Budget float64 `json:"budget"`
	Target float64 `json:"target"`
	Remain float64 `json:"remain"`
}

// SourceYearEntry is one row in the per-year, per-source, per-fund-type breakdown.
type SourceYearEntry struct {
	Year     int     `json:"year"`
	Source   string  `json:"source"`
	FundType string  `json:"fund_type"`
	Budget   float64 `json:"budget"`
	Target   float64 `json:"target"`
	Remain   float64 `json:"remain"`
}

type Snapshot struct {
	ID        int       `json:"id"`
	Label     string    `json:"label"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type SnapshotDetail struct {
	Snapshot
	Data []FlatProject `json:"data"`
}

type Scenario struct {
	ID        int       `json:"id"`
	Label     string    `json:"label"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChangeLogEntry struct {
	ID        int       `json:"id"`
	TableName string    `json:"table_name"`
	RowID     int       `json:"row_id"`
	ProjectID int       `json:"project_id"`
	RowName   string    `json:"row_name"`
	FundType  string    `json:"fund_type"`
	DataYear  int       `json:"data_year"`
	Field     string    `json:"field"`
	OldValue  float64   `json:"old_value"`
	NewValue  float64   `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}

// SubJobYearEntry is one row in the per-year, per-sub-job, per-fund-type breakdown.
type SubJobYearEntry struct {
	Name      string  `json:"name"`
	SortOrder *int    `json:"sort_order"`
	Year      int     `json:"year"`
	FundType  string  `json:"fund_type"`
	Budget    float64 `json:"budget"`
	Target    float64 `json:"target"`
	Remain    float64 `json:"remain"`
}

type FlatProject struct {
	ID              int               `json:"id"`
	ProjectCode     string            `json:"project_code"`
	ItemNo          *string           `json:"item_no"`
	Name            string            `json:"name"`
	Division        *string           `json:"division"`
	ProjectType     string            `json:"project_type"`
	Year            int               `json:"year"`
	SubJobs         []SubJobYearEntry `json:"sub_jobs"`
	SourceBreakdown []SourceYearEntry `json:"source_breakdown"`
}

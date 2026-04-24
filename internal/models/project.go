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

package models

import "encoding/json"

// ReportGroup and Report back the report-editor feature (app/report/* in the frontend). Their
// JSON field names are camelCase (groupId, presetId) rather than this codebase's usual
// snake_case, matching the existing frontend types in lib/reportTypes.ts as-is — Data is stored
// and returned as an opaque JSONB blob, never modeled field-by-field in Go.
type ReportGroup struct {
	ID    int    `json:"id,string"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type Report struct {
	ID       int             `json:"id,string"`
	GroupID  int             `json:"groupId,string"`
	PresetID *string         `json:"presetId"`
	Data     json.RawMessage `json:"data"`
}

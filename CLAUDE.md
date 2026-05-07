# budget-app-api

Go REST API for the budget app.

## Stack
- Go 1.26, chi router, pgx/v5 (PostgreSQL 16), Docker
- Module: `github.com/domonkitti/budget-app-api`

## Run locally
```bash
docker compose up db
go run ./cmd/server
```
Set `MOCK=true` to run with `mock_data.json` instead of Postgres.

## Project code format
`I{year}{type}{seq}` — e.g. `I2570Y001`  
type: Y=รายปี, C=แผนงาน, L=สัญญาเช่า

## File layout
```
cmd/server/main.go              ← router setup, all routes registered here
internal/
  models/project.go             ← all types (Project, SubJob, BudgetSource, Snapshot, Scenario, ChangeLogEntry, ...)
  handlers/
    projects.go                 ← List, Get, Flat, CreateSubJob, CreateBudgetSource, UpdateSubJob, UpdateBudgetSource, BatchSave
    summary.go                  ← Summarize, TopN
    tags.go                     ← categories, values, project-tags, sub-job-tags, allocation-selections, SummaryByTag
    snapshots.go                ← List, Create, Get, Delete, Promote
    scenarios.go                ← List, Create, Delete, Promote, Flat, GetProject, UpdateSubJob, UpdateBudgetSource
    changelog.go                ← ListByProject, Undo, UpdateBatchComment
    meta.go                     ← FilterOptions
    mock.go                     ← mock implementations of all handlers
  db/
    postgres.go                 ← connection pool setup
    migrations.go               ← auto-runs migrations/ folder on startup
migrations/                     ← SQL files, numbered 001–010
```

## All API routes (`/api/v1/...`)
```
GET    /projects                        ← filter: year, type, division, fund_type
GET    /projects/flat
GET    /projects/{code}
POST   /sub-jobs
POST   /budget-sources
PUT    /sub-jobs/{id}
PUT    /budget-sources/{id}
POST   /batch-save
GET    /projects/{code}/history
POST   /change-log/{id}/undo
PATCH  /change-log/batch/{batchId}

GET    /filter-options
GET    /summary                         ← by=division|project_type|source
GET    /summary/top                     ← limit=10
GET    /summary/by-tag                  ← category=...

GET    /tag-categories
POST   /tag-categories
DELETE /tag-categories/{id}
GET    /tag-categories/{catID}/values
POST   /tag-categories/{catID}/values
PUT    /tag-values/{id}
DELETE /tag-values/{id}
GET    /project-tags
PUT    /project-tags
GET    /sub-job-tags
PUT    /sub-job-tags
GET    /allocation-selections
PUT    /allocation-selections

GET    /snapshots
POST   /snapshots
GET    /snapshots/{id}
DELETE /snapshots/{id}
POST   /snapshots/{id}/promote

GET    /scenarios
POST   /scenarios
DELETE /scenarios/{id}
POST   /scenarios/{id}/promote
GET    /scenarios/{id}/flat
GET    /scenarios/{id}/projects/{code}
PUT    /scenarios/{id}/sub-jobs/{sjID}
PUT    /scenarios/{id}/budget-sources/{bsID}
```

## Key data model
- `Project` — project_code, year, project_type, name, division, department
- `SubJob` — belongs to project; has fund_type, data_year, budget, target, remain, cut_transfer, under_budget
- `BudgetSource` — same shape as SubJob but grouped by source string
- `remain` = budget − target (computed at query time, not stored)
- `fund_type` is always: `ผูกพัน` | `ลงทุน`
- `Snapshot` — frozen copy of all project data at a point in time; Promote overwrites live data
- `Scenario` — editable sandbox copy; same Promote mechanic
- `TagCategory` / `TagValue` — flexible tag system; ProjectTag and SubJobTag assign % allocations
- `ChangeLogEntry` — audit trail for budget/target edits; supports undo and batch comment

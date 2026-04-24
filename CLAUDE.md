# budget-app-api

Go REST API for the budget app.

## Stack
- Go 1.26
- chi router
- pgx/v5 (PostgreSQL)
- Docker + PostgreSQL 16

## Project code format
`I{year}{type}{seq}` — e.g. I2570Y001
- I = Investment (fixed)
- type: Y=รายปี, C=แผนงาน, L=สัญญาเช่า
- seq = 3-digit zero-padded

## Run locally
```bash
cp .env.example .env
docker compose up db      # start postgres only
go run ./cmd/server       # run api
```

## API routes
- GET /api/v1/projects?year=&type=&division=&fund_type=
- GET /api/v1/projects/{code}
- GET /api/v1/summary?by=division|project_type|source&year=&fund_type=&source=
- GET /api/v1/summary/top?limit=10&year=&fund_type=&source=

## DB schema
3 tables: projects, sub_jobs, budget_sources
fund_type is always a row value: ผูกพัน | ลงทุน
remain = budget - target (computed at query time, not stored)

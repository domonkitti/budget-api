package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/domonkitti/budget-app-api/internal/db"
	"github.com/domonkitti/budget-app-api/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type projectRoutes interface {
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Flat(w http.ResponseWriter, r *http.Request)
	CreateSubJob(w http.ResponseWriter, r *http.Request)
	CreateBudgetSource(w http.ResponseWriter, r *http.Request)
	UpdateSubJob(w http.ResponseWriter, r *http.Request)
	UpdateBudgetSource(w http.ResponseWriter, r *http.Request)
	BatchSave(w http.ResponseWriter, r *http.Request)
}
type summaryRoutes interface {
	Summarize(w http.ResponseWriter, r *http.Request)
	TopN(w http.ResponseWriter, r *http.Request)
}
type metaRoutes interface {
	FilterOptions(w http.ResponseWriter, r *http.Request)
}
type tagRoutes interface {
	ListCategories(w http.ResponseWriter, r *http.Request)
	CreateCategory(w http.ResponseWriter, r *http.Request)
	DeleteCategory(w http.ResponseWriter, r *http.Request)
	ListValues(w http.ResponseWriter, r *http.Request)
	CreateValue(w http.ResponseWriter, r *http.Request)
	UpdateValue(w http.ResponseWriter, r *http.Request)
	DeleteValue(w http.ResponseWriter, r *http.Request)
	GetProjectTags(w http.ResponseWriter, r *http.Request)
	SetProjectTags(w http.ResponseWriter, r *http.Request)
	GetSubJobTags(w http.ResponseWriter, r *http.Request)
	SetSubJobTags(w http.ResponseWriter, r *http.Request)
	GetAllocationSelections(w http.ResponseWriter, r *http.Request)
	SetAllocationSelections(w http.ResponseWriter, r *http.Request)
	SummaryByTag(w http.ResponseWriter, r *http.Request)
}
type snapshotRoutes interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Promote(w http.ResponseWriter, r *http.Request)
}
type scenarioRoutes interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Promote(w http.ResponseWriter, r *http.Request)
	Flat(w http.ResponseWriter, r *http.Request)
	GetProject(w http.ResponseWriter, r *http.Request)
	UpdateSubJob(w http.ResponseWriter, r *http.Request)
	UpdateBudgetSource(w http.ResponseWriter, r *http.Request)
}
type changeLogRoutes interface {
	ListByProject(w http.ResponseWriter, r *http.Request)
	Undo(w http.ResponseWriter, r *http.Request)
	UpdateBatchComment(w http.ResponseWriter, r *http.Request)
}

func main() {
	godotenv.Load()

	var (
		projects   projectRoutes
		summary    summaryRoutes
		tags       tagRoutes
		meta       metaRoutes
		snapshots  snapshotRoutes
		scenarios  scenarioRoutes
		changeLogs changeLogRoutes
	)

	if os.Getenv("MOCK") == "true" {
		mockFile := os.Getenv("MOCK_FILE")
		if mockFile == "" {
			mockFile = "mock_data.json"
		}
		s := handlers.LoadMockStore(mockFile)
		projects = handlers.NewMockProjectHandler(s)
		summary = handlers.NewMockSummaryHandler(s)
		tags = handlers.NewMockTagHandler(s)
		meta = handlers.NewMockMetaHandler(s)
		snapshots = handlers.NewMockSnapshotHandler()
		scenarios = handlers.NewMockScenarioHandler()
		changeLogs = handlers.NewMockChangeLogHandler()
	} else {
		ctx := context.Background()
		pool, err := db.Connect(ctx)
		if err != nil {
			log.Fatalf("db connect: %v", err)
		}
		defer pool.Close()
		if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
			log.Fatalf("db migrate: %v", err)
		}
		projects = handlers.NewProjectHandler(pool)
		summary = handlers.NewSummaryHandler(pool)
		tags = handlers.NewTagHandler(pool)
		meta = handlers.NewMetaHandler(pool)
		snapshots = handlers.NewSnapshotHandler(pool)
		scenarios = handlers.NewScenarioHandler(pool)
		changeLogs = handlers.NewChangeLogHandler(pool)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", projects.List)
		r.Get("/projects/flat", projects.Flat)
		r.Get("/projects/{code}", projects.Get)
		r.Post("/sub-jobs", projects.CreateSubJob)
		r.Post("/budget-sources", projects.CreateBudgetSource)
		r.Put("/sub-jobs/{id}", projects.UpdateSubJob)
		r.Put("/budget-sources/{id}", projects.UpdateBudgetSource)

		r.Get("/filter-options", meta.FilterOptions)

		r.Get("/summary", summary.Summarize)
		r.Get("/summary/top", summary.TopN)
		r.Get("/summary/by-tag", tags.SummaryByTag)

		r.Get("/tag-categories", tags.ListCategories)
		r.Post("/tag-categories", tags.CreateCategory)
		r.Delete("/tag-categories/{id}", tags.DeleteCategory)

		r.Get("/tag-categories/{catID}/values", tags.ListValues)
		r.Post("/tag-categories/{catID}/values", tags.CreateValue)
		r.Put("/tag-values/{id}", tags.UpdateValue)
		r.Delete("/tag-values/{id}", tags.DeleteValue)

		r.Get("/project-tags", tags.GetProjectTags)
		r.Put("/project-tags", tags.SetProjectTags)
		r.Get("/sub-job-tags", tags.GetSubJobTags)
		r.Put("/sub-job-tags", tags.SetSubJobTags)
		r.Get("/allocation-selections", tags.GetAllocationSelections)
		r.Put("/allocation-selections", tags.SetAllocationSelections)

		r.Get("/snapshots", snapshots.List)
		r.Post("/snapshots", snapshots.Create)
		r.Get("/snapshots/{id}", snapshots.Get)
		r.Delete("/snapshots/{id}", snapshots.Delete)
		r.Post("/snapshots/{id}/promote", snapshots.Promote)

		r.Get("/scenarios", scenarios.List)
		r.Post("/scenarios", scenarios.Create)
		r.Delete("/scenarios/{id}", scenarios.Delete)
		r.Post("/scenarios/{id}/promote", scenarios.Promote)
		r.Get("/scenarios/{id}/flat", scenarios.Flat)
		r.Get("/scenarios/{id}/projects/{code}", scenarios.GetProject)
		r.Put("/scenarios/{id}/sub-jobs/{sjID}", scenarios.UpdateSubJob)
		r.Put("/scenarios/{id}/budget-sources/{bsID}", scenarios.UpdateBudgetSource)

		r.Post("/batch-save", projects.BatchSave)

		r.Get("/projects/{code}/history", changeLogs.ListByProject)
		r.Post("/change-log/{id}/undo", changeLogs.Undo)
		r.Patch("/change-log/batch/{batchId}", changeLogs.UpdateBatchComment)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

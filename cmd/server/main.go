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

func main() {
	godotenv.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	projects := handlers.NewProjectHandler(pool)
	summary := handlers.NewSummaryHandler(pool)
	tags := handlers.NewTagHandler(pool)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", projects.List)
		r.Get("/projects/flat", projects.Flat)
		r.Get("/projects/{code}", projects.Get)

		r.Get("/summary", summary.Summarize)
		r.Get("/summary/top", summary.TopN)
		r.Get("/summary/by-tag", tags.SummaryByTag)

		r.Get("/tag-categories", tags.ListCategories)
		r.Post("/tag-categories", tags.CreateCategory)
		r.Delete("/tag-categories/{id}", tags.DeleteCategory)

		r.Get("/tag-categories/{catID}/values", tags.ListValues)
		r.Post("/tag-categories/{catID}/values", tags.CreateValue)
		r.Delete("/tag-values/{id}", tags.DeleteValue)

		r.Get("/sub-job-tags", tags.GetSubJobTags)
		r.Put("/sub-job-tags", tags.SetSubJobTags)
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

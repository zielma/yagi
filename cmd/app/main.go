package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	ihttp "github.com/zielma/yagi/internal/http"
	"github.com/zielma/yagi/internal/router"
	"github.com/zielma/yagi/internal/scheduler"
	"github.com/zielma/yagi/templates"
)

func main() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))

	config, err := config.NewFromEnv()
	if err != nil {
		slog.Error("failed to get config", "error", err)
		os.Exit(1)
	}

	db, err := database.Initialize()

	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	r := router.NewRouter()
	r.Group(func(r *router.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				slog.Info("middleware")
				h.ServeHTTP(w, req)
			})
		})

		// Parse all templates from the embedded filesystem
		_ = template.Must(template.ParseFS(templates.TemplatesFS,
			"base.layout.tmpl",
			"index.page.tmpl",
			"error-handler.partial.tmpl",
			"ynab-success.partial.tmpl",
		))

		r.Get("/hello", func(wr http.ResponseWriter, req *http.Request) {
			tmpl, err := template.ParseFiles("templates/base.layout.tmpl", "templates/index.page.tmpl")
			if err != nil {
				http.Error(wr, err.Error(), http.StatusInternalServerError)
				return
			}

			wr.Header().Set("Content-Type", "text/html")

			err = tmpl.ExecuteTemplate(wr, "index.page.tmpl", nil)
			if err != nil {
				http.Error(wr, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	})

	scheduler.New(db, config)

	server := ihttp.NewServer(r)
	err = server.ListenAndServe()
	if err != nil {
		slog.Error("server error", slog.Any("error", err))
	}
}

package main

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	ihttp "github.com/zielma/yagi/internal/http"
	"github.com/zielma/yagi/internal/jobs"
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

	jobs.RegisterJobs()
	s, err := scheduler.New(db, config)
	if err != nil {
		slog.Error("failed to create scheduler", slog.Any("error", err))
		os.Exit(1)
	}

	if err = s.Load(); err != nil {
		slog.Error("failed to load jobs", slog.Any("error", err))
		os.Exit(1)
	}

	s.Start()
	defer s.Shutdown()

	server := ihttp.NewServer(r)
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.Any("error", err))
		}
		slog.Info("shutting down server")
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", slog.Any("error", err))
	}

	slog.Info("shutdown complete")
}

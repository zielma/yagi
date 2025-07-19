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
	"github.com/zielma/yagi/internal/router"
	"github.com/zielma/yagi/internal/scheduler"
	"github.com/zielma/yagi/internal/ynab"
	"github.com/zielma/yagi/templates"
)

func main() {
	// Set up logging
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))

	// Load the configuration from the evironment
	config, err := config.NewFromEnv()
	if err != nil {
		slog.Error("failed to get config", "error", err)
		os.Exit(1)
	}

	// Check API connections
	if ok, err := ynab.NewClient(config).CheckConnection(); !ok {
		slog.Error("failed to connect to YNAB API", "error", err)
		os.Exit(1)
	}

	// Initialize connection to the database
	db, err := database.Initialize()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Set up the HTTP server router
	r := router.New()
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

	// Set up the job scheduler
	schedulerStore := scheduler.NewStore(db)
	s, err := scheduler.New(schedulerStore, config)
	if err != nil {
		slog.Error("failed to create scheduler", slog.Any("error", err))
		os.Exit(1)
	}

	// Load jobs from the database
	if err = s.Load(); err != nil {
		slog.Error("failed to load jobs", slog.Any("error", err))
		os.Exit(1)
	}

	// Start the scheduler
	s.Start()
	defer s.Shutdown()

	// Start the HTTP server
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

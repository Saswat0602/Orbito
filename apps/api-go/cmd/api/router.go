package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/events"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// BuildRouter constructs the Chi router with all ingress middleware, global error handlers, and routes.
func BuildRouter(cfg *config.Config, db *database.DB, appLog *slog.Logger) chi.Router {
	r, _ := buildRouterWithGateway(cfg, db, appLog)
	return r
}

func buildRouterWithGateway(cfg *config.Config, db *database.DB, appLog *slog.Logger) (chi.Router, *events.Gateway) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(appLog))
	r.Use(middleware.Recoverer(cfg.IsProduction()))
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecureHeaders(cfg.IsProduction()))
	r.Use(middleware.BodyLimit(cfg.MaxFileSize))

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		httpx.RespondError(w, req, apperror.NotFound(fmt.Sprintf("Cannot %s %s", req.Method, httpx.RequestPath(req))))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		httpx.RespondError(w, req, apperror.NotFound(fmt.Sprintf("Cannot %s %s", req.Method, httpx.RequestPath(req))))
	})

	r.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondText(w, http.StatusOK, "Welcome to Orbito Backend API!")
	})

	deps := buildPlatformDeps(cfg, db, appLog)

	r.Handle("/socket.io", deps.eventsGW)
	r.Handle("/socket.io/*", deps.eventsGW)

	if deps.localStore != nil {
		r.Get("/uploads/*", func(w http.ResponseWriter, req *http.Request) {
			if err := deps.localStore.StreamLocal(w, req, chi.URLParam(req, "*"), false); err != nil {
				httpx.RespondError(w, req, err)
			}
		})
	}

	mountCoreRoutes(r, deps)
	mountTenancyRoutes(r, deps)
	mountWorkRoutes(r, deps)

	return r, deps.eventsGW
}

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/modules/activity_log"
	"github.com/orbito/api-go/internal/modules/admin"
	"github.com/orbito/api-go/internal/modules/auth"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/health"
	"github.com/orbito/api-go/internal/modules/notifications"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/modules/users"
)

func mountCoreRoutes(r chi.Router, d *platformDeps) {
	healthHandler := health.NewHandler(health.NewService(d.db, d.cfg))
	r.Mount("/api/health", health.Routes(healthHandler))

	authHandler := auth.NewHandler(d.authSvc, d.cfg.JWTSecret, d.cfg.IsProduction(), d.frontendURL)
	r.Mount("/api/auth", auth.Routes(authHandler, d.authMw))

	settingsHandler := settings.NewHandler(d.settingsSvc)
	r.Mount("/api/settings", settings.Routes(settingsHandler, d.authMw))

	emailHandler := email.NewHandler(d.emailSvc)
	r.Mount("/api/email", email.Routes(emailHandler, d.authMw))

	usersSvc := users.NewService(initRepo(d.db, users.NewRepository))
	usersSvc.SetPresenceChecker(d.eventsGW)
	r.Mount("/api/users", users.Routes(users.NewHandler(usersSvc), d.authMw))

	notifHandler := notifications.NewHandler(d.notificationsSvc)
	r.Mount("/api/notifications", notifications.Routes(notifHandler, d.authMw))

	activityHandler := activity_log.NewHandler(d.activitySvc)
	r.Mount("/api/activity-logs", activity_log.Routes(activityHandler, d.authMw))
	r.Mount("/api/activity-log", activity_log.Routes(activityHandler, d.authMw))

	adminSvc := admin.NewService(initRepo(d.db, admin.NewRepository), d.settingsSvc, d.emailSvc, d.cfg)
	r.Mount("/api/admin", admin.Routes(admin.NewHandler(adminSvc), d.authMw))
}

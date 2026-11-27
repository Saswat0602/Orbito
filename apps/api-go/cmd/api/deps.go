package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/activity_log"
	"github.com/orbito/api-go/internal/modules/auth"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/notifications"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/events"
	"github.com/orbito/api-go/internal/platform/mailer"
	"github.com/orbito/api-go/internal/platform/storage"
)

type notifEmailAdapter struct {
	emailSvc *email.Service
}

func (a *notifEmailAdapter) SendAsync(to, subject, template string, data map[string]any) {
	if a.emailSvc == nil {
		return
	}
	_ = a.emailSvc.SendEmail(context.Background(), email.SendEmailDTO{
		To:       to,
		Subject:  subject,
		Template: template,
		Data:     data,
		Priority: "HIGH",
	})
}

type platformDeps struct {
	cfg              *config.Config
	db               *database.DB
	appLog           *slog.Logger
	frontendURL      string
	settingsSvc      *settings.Service
	emailSvc         *email.Service
	authSvc          *auth.Service
	authMw           func(http.Handler) http.Handler
	rbacResolver     middleware.RBACResolver
	eventsGW         *events.Gateway
	activitySvc      activity_log.Service
	notificationsSvc notifications.Service
	commentHooks     *notifications.CommentHooks
	localStore       *storage.LocalProvider
	storeProvider    storage.Provider
}

func initRepo[T any](db *database.DB, ctor func(*database.DB) T) T {
	if db == nil {
		var zero T
		return zero
	}
	return ctor(db)
}

func resolveFrontendURL(cfg *config.Config) string {
	if len(cfg.CORSOrigins) > 0 && cfg.CORSOrigins[0] != "" {
		return cfg.CORSOrigins[0]
	}
	return "http://localhost:3001"
}

func buildPlatformDeps(cfg *config.Config, db *database.DB, appLog *slog.Logger) *platformDeps {
	frontendURL := resolveFrontendURL(cfg)
	fieldCrypto, _ := crypto.NewFieldCrypto(cfg.EncryptionKey)
	settingsSvc := settings.NewService(initRepo(db, settings.NewRepository), fieldCrypto)

	mailerClient := mailer.New(mailer.Config{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser,
		Pass: cfg.SMTPPass, From: cfg.SMTPFrom, Insecure: !cfg.IsProduction(),
	})
	emailSvc := email.NewService(mailerClient, 2)

	tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiresIn, cfg.JWTRefreshExpiresIn)
	authRepo := initRepo(db, auth.NewRepository)
	var userChecker middleware.UserChecker
	if authRepo != nil {
		userChecker = authRepo.GetUserByID
	}

	eventsGW := events.NewGateway(tokenSvc, userChecker, cfg.CORSOrigins, appLog)
	activitySvc := activity_log.NewService(initRepo(db, activity_log.NewRepository))
	notifRepo := initRepo(db, notifications.NewRepository)
	notifSvc := notifications.NewService(notifRepo, eventsGW)

	var localStore *storage.LocalProvider
	var storeProvider storage.Provider
	if ls, err := storage.NewLocalProvider(cfg.UploadDest); err == nil {
		localStore = ls
		storeProvider = ls
	} else if appLog != nil {
		appLog.Warn("failed to initialize local storage provider", "err", err)
	}

	return &platformDeps{
		cfg: cfg, db: db, appLog: appLog, frontendURL: frontendURL,
		settingsSvc: settingsSvc, emailSvc: emailSvc,
		authSvc:      auth.NewService(authRepo, tokenSvc, emailSvc, settingsSvc, frontendURL),
		authMw:       middleware.RequireAuth(tokenSvc, userChecker),
		rbacResolver: middleware.NewDBResolver(db),
		eventsGW:     eventsGW, activitySvc: activitySvc, notificationsSvc: notifSvc,
		commentHooks:  notifications.NewCommentHooks(notifSvc, notifRepo, &notifEmailAdapter{emailSvc: emailSvc}, frontendURL),
		localStore:    localStore,
		storeProvider: storeProvider,
	}
}

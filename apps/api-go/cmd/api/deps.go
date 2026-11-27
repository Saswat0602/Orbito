package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/activity_log"
	"github.com/orbito/api-go/internal/modules/auth"
	"github.com/orbito/api-go/internal/modules/automation"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/notifications"
	"github.com/orbito/api-go/internal/modules/scheduler"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/events"
	"github.com/orbito/api-go/internal/platform/mailer"
	"github.com/orbito/api-go/internal/platform/queue"
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

type automationNotifAdapter struct {
	notifSvc notifications.Service
}

func (a *automationNotifAdapter) CreateAutomationNotification(ctx context.Context, userID uuid.UUID, title, message string) error {
	if a.notifSvc == nil {
		return nil
	}
	_, err := a.notifSvc.CreateNotification(ctx, notifications.CreateInput{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    sqlc.NotificationTypeTASKASSIGNED,
	})
	return err
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
	storageMgr       *storage.Manager
	storeProvider    storage.Provider
	queueRt          *queue.Runtime
	automationSvc    automation.Service
	cronSched        *scheduler.Scheduler
}

func initRepo[T any](db *database.DB, ctor func(*database.DB) T) T {
	var zero T
	if db == nil {
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

	var storageMgr *storage.Manager
	var storeProvider storage.Provider
	if mgr, err := storage.NewManager(context.Background(), cfg, appLog); err == nil {
		storageMgr = mgr
		storeProvider = mgr
	} else if appLog != nil {
		appLog.Warn("failed to initialize storage manager", "err", err)
	}

	queueRt := queue.NewRuntime(cfg, appLog)
	autoRepo := initRepo(db, automation.NewRepository)
	automationSvc := automation.NewService(autoRepo, queueRt, eventsGW, &automationNotifAdapter{notifSvc: notifSvc}, appLog)
	if db != nil {
		automation.RegisterProcessor(queueRt, automationSvc)
	}

	var cronSched *scheduler.Scheduler
	if db != nil {
		cronSched = scheduler.New(initRepo(db, scheduler.NewRepository), notifSvc, emailSvc, appLog)
	}

	return &platformDeps{
		cfg: cfg, db: db, appLog: appLog, frontendURL: frontendURL,
		settingsSvc: settingsSvc, emailSvc: emailSvc,
		authSvc:          auth.NewService(authRepo, tokenSvc, emailSvc, settingsSvc, frontendURL),
		authMw:           middleware.RequireAuth(tokenSvc, userChecker),
		rbacResolver:     middleware.NewDBResolver(db),
		eventsGW:         eventsGW,
		activitySvc:      activitySvc,
		notificationsSvc: notifSvc,
		commentHooks:     notifications.NewCommentHooks(notifSvc, notifRepo, &notifEmailAdapter{emailSvc: emailSvc}, frontendURL),
		storageMgr:       storageMgr,
		storeProvider:    storeProvider,
		queueRt:          queueRt,
		automationSvc:    automationSvc,
		cronSched:        cronSched,
	}
}

func (d *platformDeps) StartBackground(ctx context.Context) {
	if d.queueRt != nil {
		if err := d.queueRt.Start(ctx); err != nil && d.appLog != nil {
			d.appLog.Warn("Failed to start queue runtime", "err", err)
		}
	}
	if d.cronSched != nil {
		if err := d.cronSched.Start(); err != nil && d.appLog != nil {
			d.appLog.Warn("Failed to start cron scheduler", "err", err)
		}
	}
}

func (d *platformDeps) Close() {
	if d.cronSched != nil {
		d.cronSched.Stop()
	}
	if d.queueRt != nil {
		_ = d.queueRt.Close()
	}
	if d.emailSvc != nil {
		d.emailSvc.Close()
	}
	if d.eventsGW != nil {
		d.eventsGW.Close()
	}
}

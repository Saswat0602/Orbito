package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/admin"
	"github.com/orbito/api-go/internal/modules/auth"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/health"
	"github.com/orbito/api-go/internal/modules/invitations"
	"github.com/orbito/api-go/internal/modules/labels"
	"github.com/orbito/api-go/internal/modules/organization_members"
	"github.com/orbito/api-go/internal/modules/organizations"
	"github.com/orbito/api-go/internal/modules/project_members"
	"github.com/orbito/api-go/internal/modules/projects"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/modules/sprints"
	"github.com/orbito/api-go/internal/modules/task_statuses"
	"github.com/orbito/api-go/internal/modules/users"
	"github.com/orbito/api-go/internal/modules/workflows"
	"github.com/orbito/api-go/internal/modules/workspace_members"
	"github.com/orbito/api-go/internal/modules/workspaces"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/logger"
	"github.com/orbito/api-go/internal/platform/mailer"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

func loadDotEnv() {
	paths := []string{".env", "../../backend/.env", "../backend/.env", "backend/.env"}
	for _, p := range paths {
		cleaned := filepath.Clean(p)
		data, err := os.ReadFile(cleaned)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		break
	}
}

// BuildRouter constructs the Chi router with all ingress middleware, global error handlers, and routes.
func BuildRouter(cfg *config.Config, db *database.DB, appLog *slog.Logger) chi.Router {
	r := chi.NewRouter()

	// Ingress middleware pipeline
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(appLog))
	r.Use(middleware.Recoverer(cfg.IsProduction()))
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecureHeaders(cfg.IsProduction()))
	r.Use(middleware.BodyLimit(cfg.MaxFileSize))

	// Match NestJS 404 / 405 error responses
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		msg := fmt.Sprintf("Cannot %s %s", req.Method, httpx.RequestPath(req))
		httpx.RespondError(w, req, apperror.NotFound(msg))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		msg := fmt.Sprintf("Cannot %s %s", req.Method, httpx.RequestPath(req))
		httpx.RespondError(w, req, apperror.NotFound(msg))
	})

	// Root AppController route matching NestJS getHello()
	r.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		httpx.RespondText(w, http.StatusOK, "Welcome to Orbito Backend API!")
	})

	// Domain modules
	healthSvc := health.NewService(db, cfg)
	healthHandler := health.NewHandler(healthSvc)
	r.Mount("/api/health", health.Routes(healthHandler))

	fieldCrypto, _ := crypto.NewFieldCrypto(cfg.EncryptionKey)
	var settingsRepo settings.Repository
	if db != nil {
		settingsRepo = settings.NewRepository(db)
	}
	settingsSvc := settings.NewService(settingsRepo, fieldCrypto)
	settingsHandler := settings.NewHandler(settingsSvc)

	mailerClient := mailer.New(mailer.Config{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Pass:     cfg.SMTPPass,
		From:     cfg.SMTPFrom,
		Insecure: !cfg.IsProduction(),
	})
	emailSvc := email.NewService(mailerClient, 2)
	emailHandler := email.NewHandler(emailSvc)

	// Auth module & middleware
	frontendURL := "http://localhost:3001"
	if len(cfg.CORSOrigins) > 0 && cfg.CORSOrigins[0] != "" {
		frontendURL = cfg.CORSOrigins[0]
	}

	tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.JWTExpiresIn, cfg.JWTRefreshExpiresIn)
	var authRepo auth.Repository
	var userChecker middleware.UserChecker
	if db != nil {
		authRepo = auth.NewRepository(db)
		userChecker = authRepo.GetUserByID
	}
	authSvc := auth.NewService(authRepo, tokenSvc, emailSvc, settingsSvc, frontendURL)
	authHandler := auth.NewHandler(authSvc, cfg.JWTSecret, cfg.IsProduction(), frontendURL)
	authMw := middleware.RequireAuth(tokenSvc, userChecker)

	r.Mount("/api/auth", auth.Routes(authHandler, authMw))
	r.Mount("/api/settings", settings.Routes(settingsHandler, authMw))
	r.Mount("/api/email", email.Routes(emailHandler, authMw))

	var usersRepo users.Repository
	if db != nil {
		usersRepo = users.NewRepository(db)
	}
	usersSvc := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersSvc)
	r.Mount("/api/users", users.Routes(usersHandler, authMw))

	rbacResolver := middleware.NewDBResolver(db)
	var orgsRepo organizations.Repository
	if db != nil {
		orgsRepo = organizations.NewRepository(db)
	}
	orgsSvc := organizations.NewService(orgsRepo)
	orgsHandler := organizations.NewHandler(orgsSvc)
	r.Mount("/api/organizations", organizations.Routes(orgsHandler, authMw, rbacResolver))

	var orgMembersRepo organization_members.Repository
	if db != nil {
		orgMembersRepo = organization_members.NewRepository(db)
	}
	orgMembersSvc := organization_members.NewService(orgMembersRepo)
	orgMembersHandler := organization_members.NewHandler(orgMembersSvc)
	r.Mount("/api/organization-members", organization_members.Routes(orgMembersHandler, authMw))

	var wsRepo workspaces.Repository
	if db != nil {
		wsRepo = workspaces.NewRepository(db)
	}
	wsChartsSvc := workspaces.NewChartsService(wsRepo)
	wsSvc := workspaces.NewService(wsRepo, wsChartsSvc)
	wsHandler := workspaces.NewHandler(wsSvc)
	r.Mount("/api/workspaces", workspaces.Routes(wsHandler, authMw, rbacResolver))

	var wsMembersRepo workspace_members.Repository
	if db != nil {
		wsMembersRepo = workspace_members.NewRepository(db)
	}
	wsMembersSvc := workspace_members.NewService(wsMembersRepo)
	wsMembersHandler := workspace_members.NewHandler(wsMembersSvc)
	r.Mount("/api/workspace-members", workspace_members.Routes(wsMembersHandler, authMw))

	var invitationsRepo invitations.Repository
	if db != nil {
		invitationsRepo = invitations.NewRepository(db)
	}
	invitationsSvc := invitations.NewService(invitationsRepo, rbacResolver, emailSvc, cfg.FrontendURL)
	invitationsHandler := invitations.NewHandler(invitationsSvc)
	r.Mount("/api/invitations", invitations.Routes(invitationsHandler, authMw))

	var projectsRepo projects.Repository
	if db != nil {
		projectsRepo = projects.NewRepository(db)
	}
	projectsSvc := projects.NewService(projectsRepo, rbacResolver)
	projectsHandler := projects.NewHandler(projectsSvc)
	r.Mount("/api/projects", projects.Routes(projectsHandler, authMw))

	var projectMembersRepo project_members.Repository
	if db != nil {
		projectMembersRepo = project_members.NewRepository(db)
	}
	projectMembersSvc := project_members.NewService(projectMembersRepo, rbacResolver)
	projectMembersHandler := project_members.NewHandler(projectMembersSvc)
	r.Mount("/api/project-members", project_members.Routes(projectMembersHandler, authMw))

	var workflowsRepo workflows.Repository
	if db != nil {
		workflowsRepo = workflows.NewRepository(db)
	}
	workflowsSvc := workflows.NewService(workflowsRepo)
	workflowsHandler := workflows.NewHandler(workflowsSvc)
	r.Mount("/api/workflows", workflows.Routes(workflowsHandler, authMw))

	var taskStatusesRepo task_statuses.Repository
	if db != nil {
		taskStatusesRepo = task_statuses.NewRepository(db)
	}
	taskStatusesSvc := task_statuses.NewService(taskStatusesRepo)
	taskStatusesHandler := task_statuses.NewHandler(taskStatusesSvc)
	r.Mount("/api/task-statuses", task_statuses.Routes(taskStatusesHandler, authMw))

	var labelsRepo labels.Repository
	if db != nil {
		labelsRepo = labels.NewRepository(db)
	}
	labelsSvc := labels.NewService(labelsRepo)
	labelsHandler := labels.NewHandler(labelsSvc)
	r.Mount("/api/labels", labels.Routes(labelsHandler, authMw))

	var sprintsRepo sprints.Repository
	if db != nil {
		sprintsRepo = sprints.NewRepository(db)
	}
	sprintsSvc := sprints.NewService(sprintsRepo, nil)
	sprintsHandler := sprints.NewHandler(sprintsSvc)
	r.Mount("/api/sprints", sprints.Routes(sprintsHandler, authMw))

	var adminRepo admin.Repository
	if db != nil {
		adminRepo = admin.NewRepository(db)
	}
	adminSvc := admin.NewService(adminRepo, settingsSvc, emailSvc, cfg)
	adminHandler := admin.NewHandler(adminSvc)
	r.Mount("/api/admin", admin.Routes(adminHandler, authMw))

	return r
}

func main() {
	loadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger := logger.New(cfg.IsProduction())

	var db *database.DB
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err = database.New(ctx, cfg.DatabaseURL)
		cancel()
		if err != nil {
			appLogger.Warn("Database initial connection failed, proceeding with degraded health state", slog.String("error", err.Error()))
		} else {
			defer db.Close()
		}
	}

	router := BuildRouter(cfg, db, appLogger)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Server run context with graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		appLogger.Info("Graceful shutdown initiating...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Server forced shutdown", slog.String("error", err.Error()))
		}
		serverStopCtx()
	}()

	appLogger.Info("Orbito Go API listening", slog.String("addr", addr), slog.String("env", cfg.AppEnv))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		appLogger.Error("Server listen failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	<-serverCtx.Done()
	appLogger.Info("Server stopped cleanly")
}

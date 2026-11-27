package main

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/modules/invitations"
	"github.com/orbito/api-go/internal/modules/organization_members"
	"github.com/orbito/api-go/internal/modules/organizations"
	"github.com/orbito/api-go/internal/modules/workspace_members"
	"github.com/orbito/api-go/internal/modules/workspaces"
)

func mountTenancyRoutes(r chi.Router, d *platformDeps) {
	orgsSvc := organizations.NewService(initRepo(d.db, organizations.NewRepository))
	r.Mount("/api/organizations", organizations.Routes(organizations.NewHandler(orgsSvc), d.authMw, d.rbacResolver))

	orgMembersSvc := organization_members.NewService(initRepo(d.db, organization_members.NewRepository))
	r.Mount("/api/organization-members", organization_members.Routes(organization_members.NewHandler(orgMembersSvc), d.authMw))

	wsRepo := initRepo(d.db, workspaces.NewRepository)
	wsSvc := workspaces.NewService(wsRepo, workspaces.NewChartsService(wsRepo))
	wsHandler := workspaces.NewHandler(wsSvc)
	wsHandler.SetRecentActivityProvider(func(ctx context.Context, workspaceID uuid.UUID, limit, page int) (any, error) {
		return d.activitySvc.GetRecentActivityByWorkspaceOptimized(ctx, workspaceID, limit, page)
	})
	r.Mount("/api/workspaces", workspaces.Routes(wsHandler, d.authMw, d.rbacResolver))

	wsMembersSvc := workspace_members.NewService(initRepo(d.db, workspace_members.NewRepository))
	r.Mount("/api/workspace-members", workspace_members.Routes(workspace_members.NewHandler(wsMembersSvc), d.authMw))

	invitationsSvc := invitations.NewService(initRepo(d.db, invitations.NewRepository), d.rbacResolver, d.emailSvc, d.cfg.FrontendURL)
	r.Mount("/api/invitations", invitations.Routes(invitations.NewHandler(invitationsSvc), d.authMw))
}

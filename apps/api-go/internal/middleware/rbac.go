package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Role defines system roles matching Prisma Role enum.
type Role string

// Role constants matching Prisma Role enum.
const (
	RoleSuperAdmin Role = "SUPER_ADMIN"
	RoleOwner      Role = "OWNER"
	RoleManager    Role = "MANAGER"
	RoleMember     Role = "MEMBER"
	RoleViewer     Role = "VIEWER"
)

// RoleRank maps each role to its numeric hierarchy level.
var RoleRank = map[Role]int{
	RoleSuperAdmin: 5,
	RoleOwner:      4,
	RoleManager:    3,
	RoleMember:     2,
	RoleViewer:     1,
}

// ScopeType represents the boundary of multi-tenancy access.
type ScopeType string

// Scope constants defining tenancy boundaries.
const (
	ScopeOrganization ScopeType = "ORGANIZATION"
	ScopeWorkspace    ScopeType = "WORKSPACE"
	ScopeProject      ScopeType = "PROJECT"
)

type rbacContextKey string

const (
	scopeIDKey    rbacContextKey = "rbac_scope_id"
	memberRoleKey rbacContextKey = "rbac_member_role"
)

// WithScopeID stores the resolved scope UUID in the context.
func WithScopeID(ctx context.Context, scopeID uuid.UUID) context.Context {
	return context.WithValue(ctx, scopeIDKey, scopeID)
}

// GetScopeID retrieves the resolved scope UUID from the context.
func GetScopeID(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	val, ok := ctx.Value(scopeIDKey).(uuid.UUID)
	return val, ok
}

// WithMemberRole stores the user's role in the scope in context.
func WithMemberRole(ctx context.Context, role Role) context.Context {
	return context.WithValue(ctx, memberRoleKey, role)
}

// GetMemberRole retrieves the user's role in the current scope from context.
func GetMemberRole(ctx context.Context) (Role, bool) {
	if ctx == nil {
		return "", false
	}
	val, ok := ctx.Value(memberRoleKey).(Role)
	return val, ok
}

// ProjectScopeInfo holds projection data needed for project scope validation.
type ProjectScopeInfo struct {
	ID          uuid.UUID
	Visibility  string
	WorkspaceID uuid.UUID
}

// RBACResolver queries persistence for tenant memberships and scope slugs.
type RBACResolver interface {
	GetOrgMemberRole(ctx context.Context, userID, orgID uuid.UUID) (Role, error)
	GetOrgIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	GetWorkspaceMemberRole(ctx context.Context, userID, workspaceID uuid.UUID) (Role, error)
	GetWorkspaceIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	GetProjectMemberRole(ctx context.Context, userID, projectID uuid.UUID) (Role, error)
	GetProjectScopeInfo(ctx context.Context, projectID uuid.UUID) (*ProjectScopeInfo, error)
	GetProjectScopeInfoBySlug(ctx context.Context, slug string) (*ProjectScopeInfo, error)
}

// DBResolver implements RBACResolver against sqlc Queries.
type DBResolver struct {
	queries *sqlc.Queries
}

// NewDBResolver creates an RBACResolver using a database pool.
func NewDBResolver(db *database.DB) RBACResolver {
	if db == nil {
		return nil
	}
	return &DBResolver{queries: db.Queries()}
}

// GetOrgMemberRole retrieves user role in organization.
func (r *DBResolver) GetOrgMemberRole(ctx context.Context, userID, orgID uuid.UUID) (Role, error) {
	role, err := r.queries.GetOrganizationMemberRole(ctx, sqlc.GetOrganizationMemberRoleParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return "", err
	}
	return Role(role), nil
}

// GetOrgIDBySlug resolves organization slug to UUID.
func (r *DBResolver) GetOrgIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	return r.queries.GetOrganizationIDBySlug(ctx, slug)
}

// GetWorkspaceMemberRole retrieves user role in workspace.
func (r *DBResolver) GetWorkspaceMemberRole(ctx context.Context, userID, workspaceID uuid.UUID) (Role, error) {
	role, err := r.queries.GetWorkspaceMemberRole(ctx, sqlc.GetWorkspaceMemberRoleParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return "", err
	}
	return Role(role), nil
}

// GetWorkspaceIDBySlug resolves workspace slug to UUID.
func (r *DBResolver) GetWorkspaceIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	return r.queries.GetWorkspaceIDBySlug(ctx, slug)
}

// GetProjectMemberRole retrieves user role in project.
func (r *DBResolver) GetProjectMemberRole(ctx context.Context, userID, projectID uuid.UUID) (Role, error) {
	role, err := r.queries.GetProjectMemberRole(ctx, sqlc.GetProjectMemberRoleParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return "", err
	}
	return Role(role), nil
}

// GetProjectScopeInfo retrieves project visibility and workspace ID.
func (r *DBResolver) GetProjectScopeInfo(ctx context.Context, projectID uuid.UUID) (*ProjectScopeInfo, error) {
	p, err := r.queries.GetProjectScopeInfo(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &ProjectScopeInfo{
		ID:          p.ID,
		Visibility:  string(p.Visibility),
		WorkspaceID: p.WorkspaceID,
	}, nil
}

// GetProjectScopeInfoBySlug retrieves project info by slug.
func (r *DBResolver) GetProjectScopeInfoBySlug(ctx context.Context, slug string) (*ProjectScopeInfo, error) {
	p, err := r.queries.GetProjectScopeInfoBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &ProjectScopeInfo{
		ID:          p.ID,
		Visibility:  string(p.Visibility),
		WorkspaceID: p.WorkspaceID,
	}, nil
}

// RequireGlobalRole checks that authenticated user has at least minRole globally.
func RequireGlobalRole(minRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := ctxhelper.GetUserID(r.Context())
			if userIDStr == "" {
				httpx.RespondError(w, r, apperror.Forbidden("Unauthenticated"))
				return
			}

			userRole := Role(ctxhelper.GetUserRole(r.Context()))
			if userRole == RoleSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			if RoleRank[userRole] < RoleRank[minRole] {
				httpx.RespondError(w, r, apperror.Forbidden("Insufficient privileges for this global action"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireScopeRole verifies scope membership and role hierarchy for the target scope.
func RequireScopeRole(resolver RBACResolver, minRole Role, scopeType ScopeType, idParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := ctxhelper.GetUserID(r.Context())
			if userIDStr == "" {
				httpx.RespondError(w, r, apperror.Forbidden("Unauthenticated"))
				return
			}

			userUUID, err := uuid.Parse(userIDStr)
			if err != nil {
				httpx.RespondError(w, r, apperror.Forbidden("Unauthenticated"))
				return
			}

			userRole := Role(ctxhelper.GetUserRole(r.Context()))
			if userRole == RoleSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			scopeValue := extractScopeValue(r, scopeType, idParam)
			if scopeValue == "" {
				httpx.RespondError(w, r, apperror.Forbidden("Scope id missing"))
				return
			}

			if strings.Contains(scopeValue, ",") {
				scopeValue = strings.TrimSpace(strings.Split(scopeValue, ",")[0])
			}

			resolvedID, projInfo, err := resolveScope(r.Context(), resolver, scopeType, scopeValue)
			if err != nil {
				httpx.RespondError(w, r, apperror.NotFound(getScopeNotFoundMessage(scopeType)))
				return
			}

			// Public project read-only exception for VIEWER minimum requirement
			if scopeType == ScopeProject && projInfo != nil && projInfo.Visibility == "PUBLIC" {
				method := strings.ToUpper(r.Method)
				isReadOnly := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
				if isReadOnly && RoleRank[RoleViewer] >= RoleRank[minRole] {
					ctx := WithScopeID(r.Context(), resolvedID)
					ctx = WithMemberRole(ctx, RoleViewer)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			memberRole, err := resolveMemberRole(r.Context(), resolver, scopeType, userUUID, resolvedID, projInfo)
			if err != nil || memberRole == "" {
				httpx.RespondError(w, r, apperror.Forbidden("Not a member of this scope"))
				return
			}

			if RoleRank[memberRole] < RoleRank[minRole] {
				httpx.RespondError(w, r, apperror.Forbidden("Insufficient role"))
				return
			}

			ctx := WithScopeID(r.Context(), resolvedID)
			ctx = WithMemberRole(ctx, memberRole)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOrgRole enforces an organization-level role.
func RequireOrgRole(resolver RBACResolver, minRole Role, idParam ...string) func(http.Handler) http.Handler {
	param := "id"
	if len(idParam) > 0 && idParam[0] != "" {
		param = idParam[0]
	}
	return RequireScopeRole(resolver, minRole, ScopeOrganization, param)
}

// RequireWorkspaceRole enforces a workspace-level role.
func RequireWorkspaceRole(resolver RBACResolver, minRole Role, idParam ...string) func(http.Handler) http.Handler {
	param := "id"
	if len(idParam) > 0 && idParam[0] != "" {
		param = idParam[0]
	}
	return RequireScopeRole(resolver, minRole, ScopeWorkspace, param)
}

// RequireProjectRole enforces a project-level role.
func RequireProjectRole(resolver RBACResolver, minRole Role, idParam ...string) func(http.Handler) http.Handler {
	param := "id"
	if len(idParam) > 0 && idParam[0] != "" {
		param = idParam[0]
	}
	return RequireScopeRole(resolver, minRole, ScopeProject, param)
}

func extractScopeValue(r *http.Request, scopeType ScopeType, idParam string) string {
	// 1. Path param by explicit name
	if idParam != "" {
		if val := chi.URLParam(r, idParam); val != "" {
			return val
		}
	}

	// 2. Query param by explicit name
	if idParam != "" {
		if val := r.URL.Query().Get(idParam); val != "" {
			return val
		}
	}

	// 3. Fallback path/query params by scope convention
	var candidates []string
	switch scopeType {
	case ScopeOrganization:
		candidates = []string{"organizationId", "orgId", "id", "slug"}
	case ScopeWorkspace:
		candidates = []string{"workspaceId", "id", "slug"}
	case ScopeProject:
		candidates = []string{"projectId", "id", "slug"}
	}

	for _, cand := range candidates {
		if val := chi.URLParam(r, cand); val != "" {
			return val
		}
		if val := r.URL.Query().Get(cand); val != "" {
			return val
		}
	}

	// 4. Fallback to request body if JSON
	if r.Body != nil && strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			var bodyMap map[string]any
			if json.Unmarshal(bodyBytes, &bodyMap) == nil {
				if idParam != "" {
					if v, ok := bodyMap[idParam].(string); ok && v != "" {
						return v
					}
				}
				for _, cand := range candidates {
					if v, ok := bodyMap[cand].(string); ok && v != "" {
						return v
					}
				}
			}
		}
	}

	return ""
}

func resolveScope(ctx context.Context, resolver RBACResolver, scopeType ScopeType, val string) (uuid.UUID, *ProjectScopeInfo, error) {
	parsedUUID, err := uuid.Parse(val)
	isUUID := err == nil

	switch scopeType {
	case ScopeOrganization:
		if isUUID {
			return parsedUUID, nil, nil
		}
		id, err := resolver.GetOrgIDBySlug(ctx, val)
		return id, nil, err

	case ScopeWorkspace:
		if isUUID {
			return parsedUUID, nil, nil
		}
		id, err := resolver.GetWorkspaceIDBySlug(ctx, val)
		return id, nil, err

	case ScopeProject:
		if isUUID {
			info, err := resolver.GetProjectScopeInfo(ctx, parsedUUID)
			if err != nil {
				return uuid.Nil, nil, err
			}
			return parsedUUID, info, nil
		}
		info, err := resolver.GetProjectScopeInfoBySlug(ctx, val)
		if err != nil {
			return uuid.Nil, nil, err
		}
		return info.ID, info, nil

	default:
		return uuid.Nil, nil, apperror.NotFound("Unknown scope")
	}
}

func resolveMemberRole(
	ctx context.Context,
	resolver RBACResolver,
	scopeType ScopeType,
	userID, scopeID uuid.UUID,
	projInfo *ProjectScopeInfo,
) (Role, error) {
	switch scopeType {
	case ScopeOrganization:
		return resolver.GetOrgMemberRole(ctx, userID, scopeID)

	case ScopeWorkspace:
		return resolver.GetWorkspaceMemberRole(ctx, userID, scopeID)

	case ScopeProject:
		role, err := resolver.GetProjectMemberRole(ctx, userID, scopeID)
		if err == nil && role != "" {
			return role, nil
		}
		// If project visibility is INTERNAL, inherit membership from parent workspace
		if projInfo != nil && projInfo.Visibility == "INTERNAL" && projInfo.WorkspaceID != uuid.Nil {
			wsRole, wsErr := resolver.GetWorkspaceMemberRole(ctx, userID, projInfo.WorkspaceID)
			if wsErr == nil && wsRole != "" {
				return wsRole, nil
			}
		}
		return "", err

	default:
		return "", apperror.Forbidden("Unknown scope")
	}
}

func getScopeNotFoundMessage(scopeType ScopeType) string {
	switch scopeType {
	case ScopeOrganization:
		return "Organization not found"
	case ScopeWorkspace:
		return "Workspace not found"
	case ScopeProject:
		return "Project not found"
	default:
		return "Resource not found"
	}
}

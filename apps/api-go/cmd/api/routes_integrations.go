package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/modules/ai_chat"
	"github.com/orbito/api-go/internal/modules/inbox"
	"github.com/orbito/api-go/internal/modules/public"
)

func mountIntegrationRoutes(r, projectsRouter, tasksRouter chi.Router, d *platformDeps) {
	inboxHandler := inbox.NewHandler(d.inboxSvc)
	if projectsRouter != nil {
		projectsRouter.Mount("/{projectId}/inbox", inbox.ProjectInboxRoutes(inboxHandler, d.authMw))
	}
	if tasksRouter != nil {
		tasksRouter.Post("/{taskId}/comments/{commentId}/send-email", inboxHandler.SendCommentAsEmail)
	}
	r.Mount("/api/email-templates", inbox.EmailTemplatesRoutes(d.authMw))

	r.Mount("/api/ai-chat", ai_chat.Routes(ai_chat.NewHandler(d.aiChatSvc), d.authMw))

	publicHandler := public.NewHandler(d.publicSvc)
	publicRL := public.NewRateLimiter()
	r.Mount("/api/public/tasks", public.SharedTasksRoutes(publicHandler, publicRL))
	r.Mount("/api/public/workspaces", public.WorkspacesRoutes(publicHandler, publicRL))
	r.Mount("/api/public/project-tasks", public.ProjectTasksRoutes(publicHandler, publicRL))
}

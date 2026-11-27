package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/modules/gantt"
	"github.com/orbito/api-go/internal/modules/labels"
	"github.com/orbito/api-go/internal/modules/project_members"
	"github.com/orbito/api-go/internal/modules/projects"
	"github.com/orbito/api-go/internal/modules/sprints"
	"github.com/orbito/api-go/internal/modules/task_attachments"
	"github.com/orbito/api-go/internal/modules/task_comments"
	"github.com/orbito/api-go/internal/modules/task_dependencies"
	"github.com/orbito/api-go/internal/modules/task_label"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/modules/task_statuses"
	"github.com/orbito/api-go/internal/modules/task_watchers"
	"github.com/orbito/api-go/internal/modules/tasks"
	"github.com/orbito/api-go/internal/modules/time_entries"
	"github.com/orbito/api-go/internal/modules/workflows"
)

func mountWorkRoutes(r chi.Router, d *platformDeps) {
	projectsSvc := projects.NewService(initRepo(d.db, projects.NewRepository), d.rbacResolver)
	r.Mount("/api/projects", projects.Routes(projects.NewHandler(projectsSvc), d.authMw))

	projMembersSvc := project_members.NewService(initRepo(d.db, project_members.NewRepository), d.rbacResolver)
	r.Mount("/api/project-members", project_members.Routes(project_members.NewHandler(projMembersSvc), d.authMw))

	workflowsSvc := workflows.NewService(initRepo(d.db, workflows.NewRepository))
	r.Mount("/api/workflows", workflows.Routes(workflows.NewHandler(workflowsSvc), d.authMw))

	statusesSvc := task_statuses.NewService(initRepo(d.db, task_statuses.NewRepository))
	r.Mount("/api/task-statuses", task_statuses.Routes(task_statuses.NewHandler(statusesSvc), d.authMw))

	labelsSvc := labels.NewService(initRepo(d.db, labels.NewRepository))
	r.Mount("/api/labels", labels.Routes(labels.NewHandler(labelsSvc), d.authMw))

	sprintsSvc := sprints.NewService(initRepo(d.db, sprints.NewRepository), d.eventsGW)
	r.Mount("/api/sprints", sprints.Routes(sprints.NewHandler(sprintsSvc), d.authMw))

	taskRanksSvc := task_ranks.NewService(initRepo(d.db, task_ranks.NewRepository))
	r.Mount("/api/task-ranks", task_ranks.Routes(task_ranks.NewHandler(taskRanksSvc), d.authMw))

	tasksSvc := tasks.NewService(initRepo(d.db, tasks.NewRepository), taskRanksSvc, d.storeProvider, d.eventsGW, d.frontendURL)
	tasks.AttachPhase5Hooks(tasksSvc, d.activitySvc, d.notificationsSvc)
	tasksHandler := tasks.NewHandler(tasksSvc)
	r.Mount("/api/tasks", tasks.Routes(tasksHandler, d.authMw))
	r.Mount("/api/task-shares", tasks.ShareRoutes(tasksHandler, d.authMw))

	depsSvc := task_dependencies.NewService(initRepo(d.db, task_dependencies.NewRepository))
	r.Mount("/api/task-dependencies", task_dependencies.Routes(task_dependencies.NewHandler(depsSvc), d.authMw))

	watchersSvc := task_watchers.NewService(initRepo(d.db, task_watchers.NewRepository))
	r.Mount("/api/task-watchers", task_watchers.Routes(task_watchers.NewHandler(watchersSvc), d.authMw))

	taskLabelSvc := task_label.NewService(initRepo(d.db, task_label.NewRepository))
	r.Mount("/api/task-labels", task_label.Routes(task_label.NewHandler(taskLabelSvc), d.authMw))

	commentsSvc := task_comments.NewService(initRepo(d.db, task_comments.NewRepository), d.eventsGW, d.commentHooks)
	r.Mount("/api/task-comments", task_comments.Routes(task_comments.NewHandler(commentsSvc), d.authMw))

	attachSvc := task_attachments.NewService(initRepo(d.db, task_attachments.NewRepository), d.storeProvider)
	r.Mount("/api/task-attachments", task_attachments.Routes(task_attachments.NewHandler(attachSvc), d.authMw))

	timeSvc := time_entries.NewService(initRepo(d.db, time_entries.NewRepository))
	time_entries.AttachEventEmitter(timeSvc, d.eventsGW)
	r.Mount("/api/time-entries", time_entries.Routes(time_entries.NewHandler(timeSvc), d.authMw))

	ganttSvc := gantt.NewService(initRepo(d.db, gantt.NewRepository))
	r.Mount("/api/gantt", gantt.Routes(gantt.NewHandler(ganttSvc), d.authMw))
}

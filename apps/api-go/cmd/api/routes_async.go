package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/modules/automation"
	"github.com/orbito/api-go/internal/modules/editor_images"
	"github.com/orbito/api-go/internal/modules/search"
	"github.com/orbito/api-go/internal/modules/storage_files"
	"github.com/orbito/api-go/internal/platform/queue"
)

func mountAsyncRoutes(r chi.Router, d *platformDeps) {
	if d.queueRt != nil {
		r.Get("/api/health/queue", queue.HealthHandler(d.queueRt))
		r.Mount("/api/queues", queue.DashboardRoutes(d.queueRt, d.cfg.BullBoardEmail, d.cfg.BullBoardPassword))
	}

	if d.storageMgr != nil {
		filesHandler := storage_files.NewHandler(d.storageMgr, d.cfg)
		s3Handler := storage_files.NewS3Handler(d.storageMgr)
		r.Mount("/api/uploads", storage_files.UploadRoutes(filesHandler, d.authMw))
		r.Mount("/api/s3", storage_files.S3Routes(s3Handler, d.authMw))
	}

	editorSvc := editor_images.NewService(initRepo(d.db, editor_images.NewRepository), d.storageMgr, d.cfg)
	r.Mount("/api/editor-images", editor_images.Routes(editor_images.NewHandler(editorSvc), d.authMw))

	r.Mount("/api/automation", automation.Routes(automation.NewHandler(d.automationSvc), d.authMw))

	searchSvc := search.NewService(initRepo(d.db, search.NewRepository))
	r.Mount("/api/search", search.Routes(search.NewHandler(searchSvc), d.authMw))
}

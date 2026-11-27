-- name: GetProjectByIDForInbox :one
SELECT id, slug, name, workspace_id FROM "projects" WHERE id = $1 LIMIT 1;

-- name: GetProjectInboxByProjectID :one
SELECT
    pi.id, pi.project_id, pi.enabled, pi.name, pi.description, pi.email_address,
    pi.email_signature, pi.auto_reply_enabled, pi.auto_reply_template, pi.auto_create_task,
    pi.default_task_type, pi.default_priority, pi.default_status_id, pi.default_assignee_id,
    pi.sync_interval, pi.created_at, pi.updated_at,
    p.slug AS project_slug, p.name AS project_name,
    ts.name AS default_status_name, ts.color AS default_status_color,
    da.id AS assignee_id, da.first_name AS assignee_first_name, da.last_name AS assignee_last_name, da.email AS assignee_email,
    (SELECT COUNT(*)::bigint FROM "inbox_messages" im WHERE im.project_inbox_id = pi.id) AS messages_count,
    (SELECT COUNT(*)::bigint FROM "inbox_rules" ir WHERE ir.project_inbox_id = pi.id) AS rules_count
FROM "project_inboxes" pi
JOIN "projects" p ON pi.project_id = p.id
LEFT JOIN "task_statuses" ts ON pi.default_status_id = ts.id
LEFT JOIN "users" da ON pi.default_assignee_id = da.id
WHERE pi.project_id = $1
LIMIT 1;

-- name: GetProjectInboxByEmailAddress :one
SELECT id, project_id, email_address FROM "project_inboxes" WHERE email_address = $1 LIMIT 1;

-- name: CreateProjectInboxRecord :one
INSERT INTO "project_inboxes" (
    id, project_id, enabled, name, description, email_address, email_signature,
    auto_reply_enabled, auto_reply_template, auto_create_task, default_task_type,
    default_priority, default_status_id, default_assignee_id, sync_interval,
    created_at, updated_at
) VALUES (
    $1, $2, true, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $13, $14,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateProjectInboxRecord :one
UPDATE "project_inboxes"
SET
    name = COALESCE(sqlc.narg(name)::text, name),
    description = COALESCE(sqlc.narg(description)::text, description),
    email_address = COALESCE(sqlc.narg(email_address)::text, email_address),
    email_signature = COALESCE(sqlc.narg(email_signature)::text, email_signature),
    auto_reply_enabled = COALESCE(sqlc.narg(auto_reply_enabled)::boolean, auto_reply_enabled),
    auto_reply_template = COALESCE(sqlc.narg(auto_reply_template)::text, auto_reply_template),
    auto_create_task = COALESCE(sqlc.narg(auto_create_task)::boolean, auto_create_task),
    default_task_type = COALESCE(sqlc.narg(default_task_type)::"TaskType", default_task_type),
    default_priority = COALESCE(sqlc.narg(default_priority)::"TaskPriority", default_priority),
    default_status_id = COALESCE(sqlc.narg(default_status_id)::uuid, default_status_id),
    default_assignee_id = COALESCE(sqlc.narg(default_assignee_id)::uuid, default_assignee_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)::uuid
RETURNING *;

-- name: UpdateInboxEmailAddressByID :exec
UPDATE "project_inboxes" SET email_address = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: GetEmailAccountByInboxID :one
SELECT * FROM "email_accounts" WHERE project_inbox_id = $1 LIMIT 1;

-- name: GetEmailAccountByID :one
SELECT * FROM "email_accounts" WHERE id = $1 LIMIT 1;

-- name: UpsertEmailAccountForInbox :one
INSERT INTO "email_accounts" (
    id, project_inbox_id, email_address, display_name,
    imap_host, imap_port, imap_username, imap_password, imap_use_ssl,
    imap_tls_reject_unauthorized, imap_tls_min_version, imap_servername, imap_folder,
    smtp_host, smtp_port, smtp_username, smtp_password,
    smtp_tls_reject_unauthorized, smtp_tls_min_version, smtp_servername, smtp_require_tls,
    sync_enabled, created_at, updated_at
) VALUES (
    sqlc.arg(id)::uuid, sqlc.arg(project_inbox_id)::uuid, sqlc.arg(email_address)::text, sqlc.narg(display_name)::text,
    sqlc.narg(imap_host)::text, sqlc.narg(imap_port)::integer, sqlc.narg(imap_username)::text, sqlc.narg(imap_password)::text, sqlc.arg(imap_use_ssl)::boolean,
    sqlc.arg(imap_tls_reject_unauthorized)::boolean, sqlc.narg(imap_tls_min_version)::text, sqlc.narg(imap_servername)::text, sqlc.arg(imap_folder)::text,
    sqlc.narg(smtp_host)::text, sqlc.narg(smtp_port)::integer, sqlc.narg(smtp_username)::text, sqlc.narg(smtp_password)::text,
    sqlc.arg(smtp_tls_reject_unauthorized)::boolean, sqlc.narg(smtp_tls_min_version)::text, sqlc.narg(smtp_servername)::text, sqlc.arg(smtp_require_tls)::boolean,
    true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT (project_inbox_id) DO UPDATE SET
    email_address = EXCLUDED.email_address,
    display_name = EXCLUDED.display_name,
    imap_host = EXCLUDED.imap_host,
    imap_port = EXCLUDED.imap_port,
    imap_username = EXCLUDED.imap_username,
    imap_password = COALESCE(EXCLUDED.imap_password, "email_accounts".imap_password),
    imap_use_ssl = EXCLUDED.imap_use_ssl,
    imap_tls_reject_unauthorized = EXCLUDED.imap_tls_reject_unauthorized,
    imap_tls_min_version = EXCLUDED.imap_tls_min_version,
    imap_servername = EXCLUDED.imap_servername,
    imap_folder = EXCLUDED.imap_folder,
    smtp_host = EXCLUDED.smtp_host,
    smtp_port = EXCLUDED.smtp_port,
    smtp_username = EXCLUDED.smtp_username,
    smtp_password = COALESCE(EXCLUDED.smtp_password, "email_accounts".smtp_password),
    smtp_tls_reject_unauthorized = EXCLUDED.smtp_tls_reject_unauthorized,
    smtp_tls_min_version = EXCLUDED.smtp_tls_min_version,
    smtp_servername = EXCLUDED.smtp_servername,
    smtp_require_tls = EXCLUDED.smtp_require_tls,
    sync_enabled = true,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListSyncEnabledEmailAccounts :many
SELECT
    ea.*,
    pi.project_id, pi.enabled AS inbox_enabled, pi.sync_interval, pi.default_status_id,
    pi.default_task_type, pi.default_priority, pi.default_assignee_id, pi.auto_create_task,
    pi.auto_reply_enabled, pi.auto_reply_template, pi.name AS inbox_name, pi.email_signature,
    p.slug AS project_slug, p.name AS project_name, p.workspace_id, w.organization_id
FROM "email_accounts" ea
JOIN "project_inboxes" pi ON ea.project_inbox_id = pi.id
JOIN "projects" p ON pi.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE ea.sync_enabled = true
  AND pi.enabled = true;

-- name: GetEmailAccountSyncDetailsByProjectID :one
SELECT
    ea.*,
    pi.project_id, pi.enabled AS inbox_enabled, pi.sync_interval, pi.default_status_id,
    pi.default_task_type, pi.default_priority, pi.default_assignee_id, pi.auto_create_task,
    pi.auto_reply_enabled, pi.auto_reply_template, pi.name AS inbox_name, pi.email_signature,
    p.slug AS project_slug, p.name AS project_name, p.workspace_id, w.organization_id
FROM "email_accounts" ea
JOIN "project_inboxes" pi ON ea.project_inbox_id = pi.id
JOIN "projects" p ON pi.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE pi.project_id = $1
LIMIT 1;

-- name: UpdateEmailAccountSyncState :exec
UPDATE "email_accounts"
SET last_sync_at = sqlc.arg(last_sync_at)::timestamp,
    last_sync_error = sqlc.narg(last_sync_error)::text,
    uid_next = COALESCE(sqlc.narg(uid_next)::integer, uid_next),
    uid_validity = COALESCE(sqlc.narg(uid_validity)::text, uid_validity),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)::uuid;

-- name: ListInboxMessagesFiltered :many
SELECT
    im.*,
    t.id AS task_id, t.title AS task_title, t.task_number AS task_number,
    ts.name AS task_status_name, ts.color AS task_status_color,
    u.id AS conv_user_id, u.first_name AS conv_user_first_name, u.last_name AS conv_user_last_name
FROM "inbox_messages" im
LEFT JOIN "tasks" t ON t.inbox_message_id = im.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "users" u ON im.converted_by = u.id
WHERE im.project_inbox_id = sqlc.arg(project_inbox_id)::uuid
  AND (sqlc.narg(status)::"MessageStatus" IS NULL OR im.status = sqlc.narg(status)::"MessageStatus")
  AND (sqlc.arg(include_spam)::boolean = true OR im.is_spam = false)
  AND (sqlc.narg(from_email)::text IS NULL OR im.from_email ILIKE '%' || sqlc.narg(from_email)::text || '%' ESCAPE '\')
  AND (sqlc.narg(from_date)::timestamp IS NULL OR im.email_date >= sqlc.narg(from_date)::timestamp)
  AND (sqlc.narg(to_date)::timestamp IS NULL OR im.email_date <= sqlc.narg(to_date)::timestamp)
  AND (sqlc.narg(search_text)::text IS NULL OR (
      im.subject ILIKE '%' || sqlc.narg(search_text)::text || '%' ESCAPE '\'
      OR COALESCE(im.body_text, '') ILIKE '%' || sqlc.narg(search_text)::text || '%' ESCAPE '\'
  ))
ORDER BY im.email_date DESC;

-- name: ListMessageAttachmentsBatch :many
SELECT id, message_id, filename, mime_type, size, storage_url
FROM "message_attachments"
WHERE message_id = ANY($1::uuid[]);

-- name: GetInboxMessageDetailByID :one
SELECT
    im.*,
    pi.project_id, pi.default_task_type, pi.default_priority, pi.default_status_id, pi.default_assignee_id,
    p.slug AS project_slug, p.workspace_id, w.organization_id,
    t.id AS task_id, t.title AS task_title, t.slug AS task_slug,
    u.id AS conv_user_id, u.first_name AS conv_user_first_name, u.last_name AS conv_user_last_name
FROM "inbox_messages" im
JOIN "project_inboxes" pi ON im.project_inbox_id = pi.id
JOIN "projects" p ON pi.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "tasks" t ON t.inbox_message_id = im.id
LEFT JOIN "users" u ON im.converted_by = u.id
WHERE im.id = $1
LIMIT 1;

-- name: FindInboxMessageByMessageID :one
SELECT id FROM "inbox_messages" WHERE project_inbox_id = $1 AND message_id = $2 LIMIT 1;

-- name: InsertInboxMessageRecord :one
INSERT INTO "inbox_messages" (
    id, project_inbox_id, message_id, imap_uid, thread_id, in_reply_to, "references",
    subject, from_email, from_name, to_emails, cc_emails, bcc_emails, reply_to,
    body_text, body_html, snippet, headers, has_attachments, converted, status,
    email_date, received_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, false, 'PENDING',
    $20, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: FindTaskByEmailThreadID :one
SELECT id, project_id, title, slug FROM "tasks" WHERE email_thread_id = $1 LIMIT 1;

-- name: InsertCommentFromEmailMessage :one
INSERT INTO "task_comments" (
    id, task_id, author_id, content, email_message_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetLastTaskNumberByProjectForInbox :one
SELECT COALESCE(MAX(task_number), 0)::integer FROM "tasks" WHERE project_id = $1;

-- name: InsertTaskFromInboxMessage :one
INSERT INTO "tasks" (
    id, project_id, title, description, type, priority, status_id,
    email_thread_id, allow_email_replies, inbox_message_id,
    task_number, slug, created_by_id, updated_by_id,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, true, $9,
    $10, $11, $12, $12,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id, title, slug, task_number, project_id, status_id;

-- name: MarkInboxMessageConverted :exec
UPDATE "inbox_messages"
SET status = 'CONVERTED', converted = true, converted_at = CURRENT_TIMESTAMP, converted_by = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateInboxMessageStatusByID :one
UPDATE "inbox_messages"
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CopyMessageAttachmentsToTask :exec
INSERT INTO "task_attachments" (
    id, task_id, file_name, file_path, file_size, mime_type, created_at, updated_at
)
SELECT
    gen_random_uuid(), $2, ma.filename, ma.storage_path, ma.size, ma.mime_type, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "message_attachments" ma
WHERE ma.message_id = $1;

-- name: CreateInboxRuleRecord :one
INSERT INTO "inbox_rules" (
    id, project_inbox_id, name, description, priority, enabled, conditions, actions, stop_on_match,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: ListInboxRulesByInboxID :many
SELECT * FROM "inbox_rules" WHERE project_inbox_id = $1 ORDER BY priority DESC, created_at ASC;

-- name: UpdateInboxRuleByID :one
UPDATE "inbox_rules"
SET
    name = COALESCE(sqlc.narg(name)::text, name),
    description = COALESCE(sqlc.narg(description)::text, description),
    priority = COALESCE(sqlc.narg(priority)::integer, priority),
    enabled = COALESCE(sqlc.narg(enabled)::boolean, enabled),
    conditions = COALESCE(sqlc.narg(conditions)::jsonb, conditions),
    actions = COALESCE(sqlc.narg(actions)::jsonb, actions),
    stop_on_match = COALESCE(sqlc.narg(stop_on_match)::boolean, stop_on_match),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)::uuid
RETURNING *;

-- name: DeleteInboxRuleByID :execrows
DELETE FROM "inbox_rules" WHERE id = $1;

-- name: GetCommentEmailReplyContext :one
SELECT
    tc.id, tc.content, tc.sent_as_email, tc.task_id,
    t.title AS task_title, t.slug AS task_slug, t.allow_email_replies, t.email_thread_id, t.inbox_message_id, t.project_id,
    a.first_name AS author_first_name, a.last_name AS author_last_name, a.email AS author_email,
    im.message_id AS orig_message_id, im.subject AS orig_subject, im.from_email AS orig_from_email, im.cc_emails AS orig_cc_emails, im."references" AS orig_references,
    pi.email_signature,
    ea.id AS account_id, ea.email_address AS account_email, ea.display_name AS account_display_name,
    ea.smtp_host, ea.smtp_port, ea.smtp_username, ea.smtp_password, ea.smtp_servername, ea.smtp_require_tls
FROM "task_comments" tc
JOIN "tasks" t ON tc.task_id = t.id
JOIN "users" a ON tc.author_id = a.id
LEFT JOIN "inbox_messages" im ON t.inbox_message_id = im.id
LEFT JOIN "project_inboxes" pi ON im.project_inbox_id = pi.id OR pi.project_id = t.project_id
LEFT JOIN "email_accounts" ea ON ea.project_inbox_id = pi.id
WHERE tc.id = $1
LIMIT 1;

-- name: MarkCommentSentAsEmail :exec
UPDATE "task_comments"
SET sent_as_email = true,
    email_message_id = $2,
    email_recipients = $3,
    email_sent_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateEmailSyncLog :one
INSERT INTO "email_sync_logs" (
    id, project_id, started_at, completed_at, status, error, messages_processed, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP
) RETURNING *;


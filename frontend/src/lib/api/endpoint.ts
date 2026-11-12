
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3000/api';

export const SOCKET_URL =
  process.env.NEXT_PUBLIC_SOCKET_URL || 'http://localhost:3000';

export const ENDPOINTS = {
  auth: {
    login: '/auth/login',
    register: '/auth/register',
    refresh: '/auth/refresh',
    logout: '/auth/logout',
    me: '/auth/me',
    profile: '/auth/profile',
    accessControl: '/auth/access-control',
    forgotPassword: '/auth/forgot-password',
    resetPassword: '/auth/reset-password',
    verifyResetToken: '/auth/verify-reset-token',
    registrationStatus: '/auth/registration-status',
    setupStatus: '/auth/setup-status',
    setupAdmin: '/auth/setup-admin',
    setupRequired: '/auth/setup/required',
    setup: '/auth/setup',
  },

  // Users & Profiles
  users: {
    list: '/users',
    me: '/users/me',
    profile: '/users/profile',
    changePassword: '/users/change-password',
    byId: (id: string) => `/users/${id}`,
    updateStatus: (id: string) => `/users/${id}/status`,
    search: '/users/search',
    avatar: '/users/avatar',
  },

  // Organizations & Tenancy
  organizations: {
    list: '/organizations',
    create: '/organizations',
    byId: (id: string) => `/organizations/${id}`,
    bySlug: (slug: string) => `/organizations/slug/${slug}`,
    archive: (id: string) => `/organizations/${id}/archive`,
    restore: (id: string) => `/organizations/${id}/restore`,
    switch: (id: string) => `/organizations/${id}/switch`,
    members: (orgId: string) => `/organizations/${orgId}/members`,
    memberById: (orgId: string, memberId: string) =>
      `/organizations/${orgId}/members/${memberId}`,
  },

  organizationMembers: {
    list: '/organization-members',
    create: '/organization-members',
    byId: (id: string) => `/organization-members/${id}`,
    byOrg: (orgId: string) => `/organization-members/organization/${orgId}`,
  },

  // Workspaces
  workspaces: {
    list: '/workspaces',
    create: '/workspaces',
    byId: (id: string) => `/workspaces/${id}`,
    bySlug: (slug: string) => `/workspaces/by-slug/${slug}`,
    members: (workspaceId: string) => `/workspaces/${workspaceId}/members`,
    memberById: (workspaceId: string, memberId: string) =>
      `/workspaces/${workspaceId}/members/${memberId}`,
  },

  workspaceMembers: {
    list: '/workspace-members',
    create: '/workspace-members',
    byId: (id: string) => `/workspace-members/${id}`,
    byWorkspace: (workspaceId: string) =>
      `/workspace-members/workspace/${workspaceId}`,
  },

  // Projects
  projects: {
    list: '/projects',
    create: '/projects',
    byId: (id: string) => `/projects/${id}`,
    bySlug: (slug: string) => `/projects/by-slug/${slug}`,
    byOrganization: '/projects/by-organization',
    search: '/projects/search',
    searchPaginated: '/projects/search/paginated',
    archived: '/projects/archived',
    bulkHealthStats: '/projects/bulk-health-stats',
    charts: (slug: string) => `/projects/${slug}/charts`,
    sprintBurndown: (slug: string, sprintId: string) =>
      `/projects/${slug}/charts/sprint-burndown/${sprintId}`,
    members: (projectId: string) => `/projects/${projectId}/members`,
    memberById: (projectId: string, memberId: string) =>
      `/projects/${projectId}/members/${memberId}`,
  },

  projectMembers: {
    list: '/project-members',
    create: '/project-members',
    byId: (id: string) => `/project-members/${id}`,
    byProject: (projectId: string) => `/project-members/project/${projectId}`,
  },

  // Tasks & Aggregate
  tasks: {
    list: '/tasks',
    allTasks: '/tasks/all-tasks',
    grouped: '/tasks/grouped',
    create: '/tasks',
    createWithAttachments: '/tasks/with-attachments',
    bulkCreate: '/tasks/bulk-create',
    bulkImport: '/tasks/bulk-import',
    bulkStatus: '/tasks/bulk-status-update',
    bulkAssign: '/tasks/bulk-assign',
    bulkDelete: '/tasks/bulk-delete',
    today: '/tasks/today',
    byId: (idOrSlug: string) => `/tasks/${idOrSlug}`,
    share: (taskId: string) => `/tasks/${taskId}/share`,
    recurrence: (taskId: string) => `/tasks/${taskId}/recurrence`,
    completeOccurrence: (taskId: string) => `/tasks/${taskId}/complete-occurrence`,
    recurringByProject: (projectId: string) => `/tasks/recurring/project/${projectId}`,
  },

  // LexoRank Ranking
  taskRanks: {
    reorder: (taskId: string) => `/task-ranks/${taskId}/reorder`,
    rebalance: '/task-ranks/rebalance',
  },

  // Task Statuses & Workflows
  taskStatuses: {
    list: '/task-statuses',
    create: '/task-statuses',
    fromProject: '/task-statuses/from-project',
    byProject: '/task-statuses/project',
    positions: '/task-statuses/positions',
    byId: (id: string) => `/task-statuses/${id}`,
    byWorkflow: (workflowId: string) =>
      `/task-statuses/workflow/${workflowId}`,
    reorder: '/task-statuses/reorder',
    restore: (id: string) => `/task-statuses/${id}/restore`,
  },

  workflows: {
    list: '/workflows',
    create: '/workflows',
    byId: (id: string) => `/workflows/${id}`,
  },

  // Task Collaboration
  taskComments: {
    list: '/task-comments',
    create: '/task-comments',
    byId: (id: string) => `/task-comments/${id}`,
    byTask: (taskId: string) => `/task-comments/task/${taskId}`,
  },

  taskAttachments: {
    list: '/task-attachments',
    upload: '/task-attachments/upload',
    byId: (id: string) => `/task-attachments/${id}`,
  },

  taskDependencies: {
    list: '/task-dependencies',
    create: '/task-dependencies',
    byId: (id: string) => `/task-dependencies/${id}`,
    byTask: (taskId: string) => `/task-dependencies/task/${taskId}`,
  },

  taskShares: {
    create: '/task-shares',
    byTask: (taskId: string) => `/task-shares/task/${taskId}`,
    revoke: (shareId: string) => `/task-shares/${shareId}`,
    publicTask: (token: string) => `/public/tasks/${token}`,
  },

  taskWatchers: {
    byTask: (taskId: string) => `/task-watchers/${taskId}`,
    add: '/task-watchers',
    remove: (taskId: string, userId: string) =>
      `/task-watchers/${taskId}/${userId}`,
  },

  taskLabels: {
    list: '/task-labels',
    add: '/task-labels',
    remove: (taskId: string, labelId: string) =>
      `/task-labels/${taskId}/${labelId}`,
  },

  labels: {
    list: '/labels',
    create: '/labels',
    byId: (id: string) => `/labels/${id}`,
  },

  // Sprints & Dynamic Views
  sprints: {
    list: '/sprints',
    create: '/sprints',
    byId: (id: string) => `/sprints/${id}`,
    bySlug: (slug: string) => `/sprints/slug/${slug}`,
    active: (projectId: string) => `/sprints/active/${projectId}`,
    start: (id: string) => `/sprints/${id}/start`,
    complete: (id: string) => `/sprints/${id}/complete`,
    archive: (id: string) => `/sprints/${id}/archive`,
  },

  gantt: {
    project: (projectId: string) => `/gantt/project/${projectId}`,
    resources: (projectId: string) => `/gantt/project/${projectId}/resources`,
  },

  // Time Tracking
  timeEntries: {
    list: '/time-entries',
    create: '/time-entries',
    byId: (id: string) => `/time-entries/${id}`,
    byTask: (taskId: string) => `/time-entries/task/${taskId}`,
    startTimer: '/time-entries/timer/start',
    stopTimer: '/time-entries/timer/stop',
    activeTimer: '/time-entries/timer/active',
  },

  // Conversational AI & DOM Automation
  aiChat: {
    chat: '/ai-chat/chat',
    conversations: '/ai-chat/conversations',
    conversationById: (id: string) => `/ai-chat/conversations/${id}`,
    conversationMessages: (id: string) =>
      `/ai-chat/conversations/${id}/messages`,
    testConnection: '/ai-chat/test-connection',
    generateDescription: '/ai-chat/generate-description',
    clearContext: (sessionId: string) => `/ai-chat/context/${sessionId}`,
  },

  // Automation Engine
  automation: {
    rules: '/automation/rules',
    ruleById: (id: string) => `/automation/rules/${id}`,
    toggleRule: (id: string) => `/automation/rules/${id}/toggle`,
    ruleStats: (id: string) => `/automation/rules/${id}/stats`,
  },

  // Enterprise Integrations
  jiraSync: {
    connect: '/jira-sync/connect',
    validateCredentials: '/jira-sync/validate-credentials',
    importProjects: '/jira-sync/import-projects',
    status: '/jira-sync/status',
    sync: '/jira-sync/sync',
    workspace: {
      connect: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync/connect`,
      status: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync`,
      update: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync`,
      projects: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync/projects`,
      syncedProjects: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync/synced-projects`,
      projectStatuses: (workspaceId: string, projectKey: string) =>
        `/workspaces/${workspaceId}/jira-sync/projects/${projectKey}/statuses`,
      import: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync/import`,
      syncAll: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync/sync-all`,
      disconnect: (workspaceId: string) => `/workspaces/${workspaceId}/jira-sync`,
    },
  },

  trelloSync: {
    connect: '/trello-sync/connect',
    validateBoards: '/trello-sync/validate-boards',
    importBoards: '/trello-sync/import-boards',
    status: '/trello-sync/status',
    sync: '/trello-sync/sync',
    workspace: {
      connect: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync/connect`,
      status: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync`,
      update: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync`,
      projects: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync/projects`,
      boards: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync/boards`,
      import: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync/import`,
      syncAll: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync/sync-all`,
      disconnect: (workspaceId: string) => `/workspaces/${workspaceId}/trello-sync`,
    },
  },

  inbox: {
    projectInbox: (projectId: string) => `/projects/${projectId}/inbox`,
    emailAccount: (projectId: string) =>
      `/projects/${projectId}/inbox/email-account`,
    sync: (projectId: string) => `/projects/${projectId}/inbox/sync`,
    messages: (projectId: string) => `/projects/${projectId}/inbox/messages`,
    messageById: (projectId: string, messageId: string) =>
      `/projects/${projectId}/inbox/messages/${messageId}`,
    convertMessage: (projectId: string, messageId: string) =>
      `/projects/${projectId}/inbox/messages/${messageId}/convert`,
    ignoreMessage: (projectId: string, messageId: string) =>
      `/projects/${projectId}/inbox/messages/${messageId}/ignore`,
    rules: (projectId: string) => `/projects/${projectId}/inbox/rules`,
    ruleById: (projectId: string, ruleId: string) =>
      `/projects/${projectId}/inbox/rules/${ruleId}`,
    testEmail: (projectId: string, accountId: string) =>
      `/projects/${projectId}/inbox/test-email/${accountId}`,
    templates: '/inbox/email-templates',
  },

  // Notifications, Activity, Invitations & Search
  notifications: {
    list: '/notifications',
    unreadCount: '/notifications/unread-count',
    markAsRead: (id: string) => `/notifications/${id}/read`,
    markAllAsRead: '/notifications/mark-all-read',
  },

  invitations: {
    list: '/invitations',
    create: '/invitations',
    verify: (token: string) => `/invitations/verify/${token}`,
    accept: (token: string) => `/invitations/${token}/accept`,
    decline: (token: string) => `/invitations/${token}/decline`,
    cancel: (id: string) => `/invitations/${id}`,
  },

  search: {
    query: '/search',
  },

  activityLog: {
    list: '/activity-log',
    byProject: (projectId: string) => `/activity-log/project/${projectId}`,
    taskActivities: (taskId: string) => `/activity-logs/task/${taskId}/activities`,
  },

  settings: {
    list: '/settings',
    byKey: (key: string) => `/settings/${key}`,
  },

  storage: {
    upload: '/storage/upload',
    fileById: (id: string) => `/storage/files/${id}`,
  },

  health: {
    check: '/health',
  },

  admin: {
    dashboard: '/admin/dashboard',
    users: '/admin/users',
    userById: (id: string) => `/admin/users/${id}`,
    updateUserRole: (id: string) => `/admin/users/${id}/role`,
    updateUserStatus: (id: string) => `/admin/users/${id}/status`,
    deleteUser: (id: string) => `/admin/users/${id}`,
    resetUserPassword: (id: string) => `/admin/users/${id}/reset-password`,
    organizations: '/admin/organizations',
    organizationById: (id: string) => `/admin/organizations/${id}`,
    toggleOrganizationArchive: (id: string) => `/admin/organizations/${id}/archive`,
    transferOrganizationOwnership: (id: string) => `/admin/organizations/${id}/transfer-ownership`,
    config: '/admin/config',
    testSmtp: '/admin/config/test-smtp',
  },

  queues: {
    health: '/queues/health',
    ui: '/queues',
  },
} as const;

export default ENDPOINTS;


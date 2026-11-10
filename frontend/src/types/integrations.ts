export interface IntegrationStatus {
  connected: boolean;
  lastSyncAt?: string | null;
  accountEmail?: string | null;
}

export interface JiraWorkspaceStatus {
  id: string;
  workspaceId: string;
  jiraSiteUrl: string;
  hasEmail: boolean;
  hasApiToken: boolean;
  lastSyncedAt?: string | null;
  syncIntervalMinutes?: number;
  syncStatus?: 'IDLE' | 'SYNCING' | 'FAILED' | 'SUCCESS';
  errorMessage?: string | null;
  createdAt: string;
}

export interface TrelloWorkspaceStatus {
  id: string;
  workspaceId: string;
  trelloWorkspaceId?: string | null;
  hasApiKey: boolean;
  hasToken: boolean;
  lastSyncedAt?: string | null;
  syncIntervalMinutes?: number;
  syncStatus?: 'IDLE' | 'SYNCING' | 'FAILED' | 'SUCCESS';
  errorMessage?: string | null;
  createdAt: string;
}


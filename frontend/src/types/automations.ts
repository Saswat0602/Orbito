import { TriggerType, ActionType, RuleStatus } from '@orbito/types';

export { TriggerType, ActionType, RuleStatus };

export interface AutomationRule {
  id: string;
  name: string;
  description?: string | null;
  organizationId: string;
  workspaceId?: string | null;
  projectId?: string | null;
  triggerType: TriggerType;
  triggerConfig?: Record<string, unknown>;
  actionType: ActionType;
  actionConfig?: Record<string, unknown>;
  status: RuleStatus;
  executionCount?: number;
  lastTriggeredAt?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAutomationRuleRequest {
  name: string;
  description?: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  triggerType: TriggerType;
  triggerConfig?: Record<string, unknown>;
  actionType: ActionType;
  actionConfig?: Record<string, unknown>;
}

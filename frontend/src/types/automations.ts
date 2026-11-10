import type { AutomationRule, TriggerType, ActionType, RuleStatus } from '@orbito/types';

export type { AutomationRule, TriggerType, ActionType, RuleStatus };

export interface CreateAutomationRuleRequest {
  name: string;
  description?: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  trigger?: TriggerType;
  triggerType: TriggerType;
  triggerConfig?: Record<string, unknown>;
  action?: ActionType;
  actionType: ActionType;
  actionConfig?: Record<string, unknown>;
}

import type { TriggerType, ActionType, RuleStatus } from '../enums';

export interface AutomationRule {
  id: string;
  name: string;
  description?: string | null;
  organizationId?: string | null;
  workspaceId?: string | null;
  projectId?: string | null;
  trigger?: TriggerType;
  triggerType: TriggerType;
  triggerConfig?: Record<string, unknown> | null;
  action?: ActionType;
  actionType: ActionType;
  actionConfig?: Record<string, unknown> | null;
  status: RuleStatus;
  executionCount?: number;
  lastRunAt?: string | null;
  lastExecuted?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAutomationRuleDto {
  name: string;
  description?: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  trigger?: TriggerType;
  triggerType?: TriggerType;
  triggerConfig?: Record<string, unknown>;
  action?: ActionType;
  actionType?: ActionType;
  actionConfig?: Record<string, unknown>;
}

export interface UpdateAutomationRuleDto extends Partial<CreateAutomationRuleDto> {
  status?: RuleStatus;
}

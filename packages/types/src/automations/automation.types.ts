import type { TriggerType, ActionType, RuleStatus } from '../enums';

export interface AutomationRule {
  id: string;
  name: string;
  description?: string | null;
  workspaceId: string;
  trigger: TriggerType;
  triggerConfig: Record<string, unknown>;
  action: ActionType;
  actionConfig: Record<string, unknown>;
  status: RuleStatus;
  executionCount: number;
  lastRunAt?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAutomationRuleDto {
  name: string;
  description?: string;
  workspaceId: string;
  trigger: TriggerType;
  triggerConfig: Record<string, unknown>;
  action: ActionType;
  actionConfig: Record<string, unknown>;
}

export interface UpdateAutomationRuleDto extends Partial<CreateAutomationRuleDto> {
  status?: RuleStatus;
}

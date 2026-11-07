import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';

export interface AutomationRule {
  id: string;
  name: string;
  description?: string | null;
  projectId: string;
  triggerType: string;
  triggerConfig: Record<string, unknown>;
  actionType: string;
  actionConfig: Record<string, unknown>;
  isActive: boolean;
  createdAt: string;
}

export interface CreateAutomationRuleRequest {
  name: string;
  description?: string;
  projectId: string;
  triggerType: string;
  triggerConfig: Record<string, unknown>;
  actionType: string;
  actionConfig: Record<string, unknown>;
}

export const automationApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getAutomationRules: builder.query<AutomationRule[], { projectId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.automation.rules,
        params: params || {},
      }),
      providesTags: ['Automation'],
    }),
    createAutomationRule: builder.mutation<AutomationRule, CreateAutomationRuleRequest>({
      query: (body) => ({
        url: ENDPOINTS.automation.rules,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Automation'],
    }),
    toggleAutomationRule: builder.mutation<AutomationRule, string>({
      query: (id) => ({
        url: ENDPOINTS.automation.toggleRule(id),
        method: 'PATCH',
      }),
      invalidatesTags: ['Automation'],
    }),
    deleteAutomationRule: builder.mutation<void, string>({
      query: (id) => ({
        url: ENDPOINTS.automation.ruleById(id),
        method: 'DELETE',
      }),
      invalidatesTags: ['Automation'],
    }),
  }),
});

export const {
  useGetAutomationRulesQuery,
  useCreateAutomationRuleMutation,
  useToggleAutomationRuleMutation,
  useDeleteAutomationRuleMutation,
} = automationApi;


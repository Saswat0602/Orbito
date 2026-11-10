import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';
import type { AutomationRule, CreateAutomationRuleRequest } from '@/types';

export type { AutomationRule, CreateAutomationRuleRequest };

export const automationApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getAutomationRules: builder.query<
      AutomationRule[],
      { organizationId?: string; workspaceId?: string; projectId?: string } | void
    >({
      query: (params) => ({
        url: ENDPOINTS.automation.rules,
        params: params || {},
      }),
      transformResponse: (response: AutomationRule[] | unknown) => {
        if (Array.isArray(response)) return response;
        return [];
      },
      providesTags: ['Automation'],
    }),

    getAutomationRuleById: builder.query<AutomationRule, string>({
      query: (id) => ENDPOINTS.automation.ruleById(id),
      providesTags: (_result, _error, id) => [{ type: 'Automation', id }],
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
  useGetAutomationRuleByIdQuery,
  useCreateAutomationRuleMutation,
  useToggleAutomationRuleMutation,
  useDeleteAutomationRuleMutation,
} = automationApi;

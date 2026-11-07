/**
 * Common Pagination and API Types
 */

export interface ApiResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  statusCode?: number;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  search?: string;
}

export interface PaginatedResponse<T = unknown> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
  hasNextPage?: boolean;
  hasPrevPage?: boolean;
}

export interface ColumnConfig {
  id: string;
  label: string;
  visible: boolean;
  width?: number;
  type?: string;
}

export interface DynamicColumn {
  id: string;
  label: string;
  type: string;
  icon?: unknown;
}

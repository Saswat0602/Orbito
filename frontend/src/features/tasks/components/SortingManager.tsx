'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  RotateCcw,
  FileText,
  Flame,
  CircleDot,
  CalendarDays,
  CalendarClock,
  Hash,
  Key,
  Check,
} from 'lucide-react';
import type { SortField, SortOrder } from '../types/taskViewTypes';

interface SortFieldOption {
  value: SortField;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

export const SORT_FIELD_OPTIONS: SortFieldOption[] = [
  { value: 'title', label: 'Title', icon: FileText },
  { value: 'priority', label: 'Priority', icon: Flame },
  { value: 'status', label: 'Status', icon: CircleDot },
  { value: 'dueDate', label: 'Due Date', icon: CalendarDays },
  { value: 'createdAt', label: 'Created Date', icon: CalendarClock },
  { value: 'storyPoints', label: 'Story Points', icon: Hash },
  { value: 'slug', label: 'Task Key', icon: Key },
];

interface SortingManagerProps {
  sortField: SortField;
  sortOrder: SortOrder;
  onSortFieldChange: (field: SortField) => void;
  onSortOrderChange: (order: SortOrder) => void;
  onResetSort?: () => void;
}

export const SortingManager: React.FC<SortingManagerProps> = ({
  sortField,
  sortOrder,
  onSortFieldChange,
  onSortOrderChange,
  onResetSort,
}) => {
  const activeOption = SORT_FIELD_OPTIONS.find((opt) => opt.value === sortField);

  const handleReset = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onResetSort) {
      onResetSort();
    } else {
      onSortFieldChange('createdAt');
      onSortOrderChange('desc');
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-8 border-zinc-800 bg-zinc-950/60 text-xs text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors gap-1.5 cursor-pointer ${
            sortField !== 'createdAt'
              ? 'border-indigo-500/40 bg-indigo-950/20 text-indigo-300'
              : ''
          }`}
          title="Sort tasks"
        >
          <ArrowUpDown className="h-3.5 w-3.5 text-zinc-400" />
          <span className="hidden sm:inline">
            {activeOption ? activeOption.label : 'Sort'}
          </span>
          {sortOrder === 'asc' ? (
            <ArrowUp className="h-3 w-3 text-zinc-400" />
          ) : (
            <ArrowDown className="h-3 w-3 text-zinc-400" />
          )}
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-64 bg-zinc-900 border-zinc-800 text-zinc-200 p-1.5 shadow-xl rounded-md"
      >
        <div className="flex items-center justify-between px-2 py-1.5">
          <DropdownMenuLabel className="p-0 text-xs font-semibold text-zinc-100 flex items-center gap-2">
            <ArrowUpDown className="h-3.5 w-3.5 text-zinc-400" />
            Sort Options
          </DropdownMenuLabel>
          <button
            type="button"
            onClick={handleReset}
            className="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 cursor-pointer transition-colors"
            title="Reset sorting"
          >
            <RotateCcw className="h-3 w-3" />
            Reset
          </button>
        </div>

        <div className="px-2 pb-1.5 text-[11px] text-zinc-400">
          Choose field and ordering direction
        </div>

        <DropdownMenuSeparator className="bg-zinc-800" />

        {/* Direction Switcher */}
        <div className="p-1">
          <div className="text-[10px] font-medium text-zinc-400 uppercase tracking-wider mb-1 px-1">
            Direction
          </div>
          <div className="grid grid-cols-2 gap-1">
            <button
              type="button"
              onClick={() => onSortOrderChange('asc')}
              className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded text-xs transition-colors cursor-pointer ${
                sortOrder === 'asc'
                  ? 'bg-zinc-800 text-zinc-100 font-medium border border-zinc-700'
                  : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200 border border-transparent'
              }`}
            >
              <ArrowUp className="h-3 w-3" />
              Ascending
            </button>
            <button
              type="button"
              onClick={() => onSortOrderChange('desc')}
              className={`flex items-center justify-center gap-1.5 py-1.5 px-2 rounded text-xs transition-colors cursor-pointer ${
                sortOrder === 'desc'
                  ? 'bg-zinc-800 text-zinc-100 font-medium border border-zinc-700'
                  : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200 border border-transparent'
              }`}
            >
              <ArrowDown className="h-3 w-3" />
              Descending
            </button>
          </div>
        </div>

        <DropdownMenuSeparator className="bg-zinc-800" />

        {/* Sort Fields */}
        <div className="py-1 space-y-0.5 max-h-60 overflow-y-auto">
          {SORT_FIELD_OPTIONS.map((option) => {
            const Icon = option.icon;
            const isActive = sortField === option.value;
            return (
              <DropdownMenuItem
                key={option.value}
                onClick={() => onSortFieldChange(option.value)}
                className={`flex items-center justify-between px-2.5 py-1.5 rounded text-xs cursor-pointer transition-colors ${
                  isActive
                    ? 'bg-zinc-800/80 text-zinc-100 font-medium'
                    : 'text-zinc-300 hover:bg-zinc-800/50 hover:text-zinc-100'
                }`}
              >
                <div className="flex items-center gap-2">
                  <Icon className="h-3.5 w-3.5 text-zinc-400" />
                  <span>{option.label}</span>
                </div>
                {isActive && <Check className="h-3.5 w-3.5 text-indigo-400 ml-2" />}
              </DropdownMenuItem>
            );
          })}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};


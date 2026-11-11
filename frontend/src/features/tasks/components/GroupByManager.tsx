'use client';

import * as React from 'react';
import { useCallback } from 'react';
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
  LayoutList,
  CircleDot,
  Flame,
  User,
  Shapes,
  CalendarDays,
  X,
  Check,
} from 'lucide-react';
import type { GroupByField } from '../types/taskViewTypes';

export interface GroupFieldConfig {
  value: GroupByField;
  label: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
}

export const GROUP_FIELD_OPTIONS: GroupFieldConfig[] = [
  {
    value: 'status',
    label: 'Status',
    description: 'Group tasks by workflow stage',
    icon: CircleDot,
    color: 'text-blue-400',
  },
  {
    value: 'priority',
    label: 'Priority',
    description: 'Group by urgency and severity',
    icon: Flame,
    color: 'text-orange-400',
  },
  {
    value: 'assignee',
    label: 'Assignee',
    description: 'Group by assigned team member',
    icon: User,
    color: 'text-purple-400',
  },
  {
    value: 'type',
    label: 'Task Type',
    description: 'Group by task, bug, story, or epic',
    icon: Shapes,
    color: 'text-pink-400',
  },
  {
    value: 'dueDate',
    label: 'Due Date',
    description: 'Group by overdue, today, and upcoming',
    icon: CalendarDays,
    color: 'text-emerald-400',
  },
];

export const GROUP_BY_STORAGE_KEY = 'orbito_tasks_group_by';

interface GroupByManagerProps {
  groupBy: GroupByField;
  onGroupByChange: (field: GroupByField) => void;
}

export const GroupByManager: React.FC<GroupByManagerProps> = ({
  groupBy,
  onGroupByChange,
}) => {
  const handleSelect = useCallback(
    (field: GroupByField) => {
      onGroupByChange(field);
      try {
        localStorage.setItem(GROUP_BY_STORAGE_KEY, field);
      } catch {
        // storage disabled or SSR
      }
    },
    [onGroupByChange]
  );

  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      handleSelect('none');
    },
    [handleSelect]
  );

  const activeConfig =
    groupBy !== 'none'
      ? GROUP_FIELD_OPTIONS.find((f) => f.value === groupBy) ?? null
      : null;

  const ActiveIcon = activeConfig?.icon;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-8 border-zinc-800 bg-zinc-950/60 text-xs text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors gap-1.5 cursor-pointer ${
            activeConfig
              ? 'border-indigo-500/40 bg-indigo-950/20 text-indigo-300'
              : ''
          }`}
          title="Group tasks by property"
        >
          {ActiveIcon ? (
            <>
              <ActiveIcon className={`h-3.5 w-3.5 ${activeConfig?.color ?? ''}`} />
              <span className="font-medium hidden sm:inline max-w-[85px] truncate">
                {activeConfig?.label}
              </span>
              <span
                role="button"
                tabIndex={0}
                aria-label="Clear group by"
                onClick={handleClear}
                onKeyDown={(e) => e.key === 'Enter' && handleClear(e as unknown as React.MouseEvent)}
                className="ml-0.5 rounded p-0.5 hover:bg-zinc-800 text-zinc-400 hover:text-zinc-200 cursor-pointer"
              >
                <X className="h-3 w-3" />
              </span>
            </>
          ) : (
            <>
              <LayoutList className="h-3.5 w-3.5 text-zinc-400" />
              <span className="hidden sm:inline">Group</span>
            </>
          )}
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-64 bg-zinc-900 border-zinc-800 text-zinc-200 p-1.5 shadow-xl rounded-md"
      >
        <div className="flex items-center justify-between px-2 py-1.5">
          <DropdownMenuLabel className="p-0 text-xs font-semibold text-zinc-100 flex items-center gap-2">
            <LayoutList className="h-3.5 w-3.5 text-zinc-400" />
            Group By
          </DropdownMenuLabel>
          {activeConfig && (
            <button
              type="button"
              onClick={() => handleSelect('none')}
              className="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 cursor-pointer transition-colors"
            >
              <X className="h-3 w-3" />
              Clear
            </button>
          )}
        </div>

        <div className="px-2 pb-1.5 text-[11px] text-zinc-400">
          Organize your tasks into structured swimlanes
        </div>

        <DropdownMenuSeparator className="bg-zinc-800" />

        <div className="py-1 space-y-0.5">
          {GROUP_FIELD_OPTIONS.map((option) => {
            const Icon = option.icon;
            const isActive = groupBy === option.value;
            return (
              <DropdownMenuItem
                key={option.value}
                onClick={() => handleSelect(option.value)}
                className={`flex items-center justify-between px-2.5 py-2 rounded text-xs cursor-pointer transition-colors ${
                  isActive
                    ? 'bg-zinc-800/80 text-zinc-100 font-medium'
                    : 'text-zinc-300 hover:bg-zinc-800/50 hover:text-zinc-100'
                }`}
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <Icon className={`h-4 w-4 shrink-0 ${option.color}`} />
                  <div className="truncate">
                    <div className="text-xs font-medium text-zinc-200 truncate">{option.label}</div>
                    <div className="text-[10px] text-zinc-400 truncate">{option.description}</div>
                  </div>
                </div>
                {isActive && <Check className="h-3.5 w-3.5 text-indigo-400 shrink-0 ml-2" />}
              </DropdownMenuItem>
            );
          })}
        </div>

        <DropdownMenuSeparator className="bg-zinc-800" />

        <DropdownMenuItem
          onClick={() => handleSelect('none')}
          className={`flex items-center justify-between px-2.5 py-2 rounded text-xs cursor-pointer transition-colors ${
            groupBy === 'none'
              ? 'bg-zinc-800/80 text-zinc-100 font-medium'
              : 'text-zinc-300 hover:bg-zinc-800/50 hover:text-zinc-100'
          }`}
        >
          <div className="flex items-center gap-2.5 min-w-0">
            <LayoutList className="h-4 w-4 text-zinc-400 shrink-0" />
            <div>
              <div className="text-xs font-medium text-zinc-200">No Grouping</div>
              <div className="text-[10px] text-zinc-400">Show as flat list or board</div>
            </div>
          </div>
          {groupBy === 'none' && <Check className="h-3.5 w-3.5 text-indigo-400 shrink-0 ml-2" />}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};


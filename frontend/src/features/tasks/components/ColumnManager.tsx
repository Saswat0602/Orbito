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
  SlidersHorizontal,
  RotateCcw,
  Check,
  Key,
  Shapes,
  FileText,
  CircleDot,
  Flame,
  Users,
  Target,
  Hash,
  CalendarDays,
  Clock,
  Repeat,
  CalendarClock,
} from 'lucide-react';
import type { ColumnConfig } from '../types/taskViewTypes';

interface ColumnManagerProps {
  columns: ColumnConfig[];
  onToggleColumn: (columnId: string) => void;
  onResetColumns: () => void;
}

const getColumnIcon = (id: string) => {
  switch (id) {
    case 'slug':
      return Key;
    case 'type':
      return Shapes;
    case 'title':
      return FileText;
    case 'status':
      return CircleDot;
    case 'priority':
      return Flame;
    case 'assignees':
      return Users;
    case 'sprint':
      return Target;
    case 'storyPoints':
      return Hash;
    case 'dueDate':
      return CalendarDays;
    case 'timeTracking':
      return Clock;
    case 'recurrence':
      return Repeat;
    case 'createdAt':
      return CalendarClock;
    default:
      return FileText;
  }
};

export const ColumnManager: React.FC<ColumnManagerProps> = ({
  columns,
  onToggleColumn,
  onResetColumns,
}) => {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-8 border-zinc-800 bg-zinc-950/60 text-xs text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 transition-colors gap-1.5 cursor-pointer"
          title="Customize visible columns"
        >
          <SlidersHorizontal className="h-3.5 w-3.5 text-zinc-400" />
          <span className="hidden sm:inline">Columns</span>
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-64 bg-zinc-900 border-zinc-800 text-zinc-200 p-1.5 shadow-xl rounded-md"
      >
        <div className="flex items-center justify-between px-2 py-1.5">
          <DropdownMenuLabel className="p-0 text-xs font-semibold text-zinc-100 flex items-center gap-2">
            <SlidersHorizontal className="h-3.5 w-3.5 text-zinc-400" />
            Manage Columns
          </DropdownMenuLabel>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onResetColumns();
            }}
            className="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 cursor-pointer transition-colors"
          >
            <RotateCcw className="h-3 w-3" />
            Reset
          </button>
        </div>

        <div className="px-2 pb-1.5 text-[11px] text-zinc-400">
          Customize which columns appear in your list view
        </div>

        <DropdownMenuSeparator className="bg-zinc-800" />

        <div className="py-1 space-y-0.5 max-h-72 overflow-y-auto">
          {columns.map((column) => {
            const Icon = getColumnIcon(column.id);
            const isFixed = column.fixed;

            return (
              <DropdownMenuItem
                key={column.id}
                onClick={(e) => {
                  if (isFixed) return;
                  e.preventDefault();
                  onToggleColumn(column.id);
                }}
                disabled={isFixed}
                className={`flex items-center justify-between px-2.5 py-1.5 rounded text-xs transition-colors select-none ${
                  isFixed
                    ? 'opacity-60 cursor-not-allowed text-zinc-400'
                    : 'cursor-pointer hover:bg-zinc-800/50 text-zinc-200'
                }`}
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <div
                    className={`h-4 w-4 rounded flex items-center justify-center border transition-colors ${
                      column.visible
                        ? 'bg-zinc-100 border-white text-zinc-950 font-bold'
                        : 'border-zinc-700 bg-zinc-800/40'
                    }`}
                  >
                    {column.visible && <Check className="h-3 w-3 stroke-[2.5]" />}
                  </div>
                  <Icon className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                  <span className="truncate">{column.label}</span>
                </div>
                {isFixed && (
                  <span className="text-[10px] uppercase font-mono text-zinc-500 ml-2">
                    Fixed
                  </span>
                )}
              </DropdownMenuItem>
            );
          })}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};


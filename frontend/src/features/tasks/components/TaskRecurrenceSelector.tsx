'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  Repeat,
  Calendar,
  X,
  Check,
  Trash2,
  ChevronDown,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  RecurrenceType,
  RecurrenceEndType,
  TaskRecurrence,
} from '@orbito/types';
import {
  useAddTaskRecurrenceMutation,
  useUpdateTaskRecurrenceMutation,
  useDeleteTaskRecurrenceMutation,
  useGetTaskRecurrenceQuery,
} from '@/lib/store/api/tasksApi';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface TaskRecurrenceSelectorProps {
  taskId?: string;
  isRecurring?: boolean;
  initialConfig?: Partial<TaskRecurrence> | null;
  onChange?: (config: Partial<TaskRecurrence> | null) => void;
  canEdit?: boolean;
}

const DAYS_OF_WEEK = [
  { value: 0, label: 'Sun' },
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' },
];

export function TaskRecurrenceSelector({
  taskId,
  isRecurring = false,
  initialConfig,
  onChange,
  canEdit = true,
}: TaskRecurrenceSelectorProps) {
  // Queries & Mutations for task detail drawer
  const { data: serverRecurrence, isLoading: isFetching } = useGetTaskRecurrenceQuery(
    taskId || '',
    { skip: !taskId || !isRecurring }
  );

  const [addRecurrence, { isLoading: isAdding }] = useAddTaskRecurrenceMutation();
  const [updateRecurrence, { isLoading: isUpdating }] = useUpdateTaskRecurrenceMutation();
  const [deleteRecurrence, { isLoading: isDeleting }] = useDeleteTaskRecurrenceMutation();

  const [isOpen, setIsOpen] = useState(false);

  // Form State
  const [recurrenceType, setRecurrenceType] = useState<RecurrenceType>(RecurrenceType.WEEKLY);
  const [interval, setInterval] = useState<number>(1);
  const [daysOfWeek, setDaysOfWeek] = useState<number[]>([1]); // Default to Monday
  const [dayOfMonth, setDayOfMonth] = useState<number>(1);
  const [endType, setEndType] = useState<RecurrenceEndType>(RecurrenceEndType.NEVER);
  const [endDate, setEndDate] = useState<string>('');
  const [occurrenceCount, setOccurrenceCount] = useState<number>(10);

  // Sync state from server or initialConfig
  useEffect(() => {
    const config = serverRecurrence || initialConfig;
    if (config) {
      if (config.recurrenceType || config.type) {
        setRecurrenceType(config.recurrenceType || config.type || RecurrenceType.WEEKLY);
      }
      if (typeof config.interval === 'number') {
        setInterval(config.interval);
      }
      if (Array.isArray(config.daysOfWeek)) {
        setDaysOfWeek(config.daysOfWeek);
      }
      if (typeof config.dayOfMonth === 'number') {
        setDayOfMonth(config.dayOfMonth);
      }
      if (config.endType) {
        setEndType(config.endType);
      }
      if (config.endDate) {
        setEndDate(config.endDate.split('T')[0] || '');
      }
      if (typeof (config.occurrenceCount || config.occurrencesCount) === 'number') {
        setOccurrenceCount((config.occurrenceCount || config.occurrencesCount) as number);
      }
    }
  }, [serverRecurrence, initialConfig]);

  const toggleDay = (day: number) => {
    setDaysOfWeek((prev) => {
      const exists = prev.includes(day);
      if (exists) {
        return prev.length > 1 ? prev.filter((d) => d !== day) : prev;
      }
      return [...prev, day].sort((a, b) => a - b);
    });
  };

  const handleSave = async () => {
    const payload: Partial<TaskRecurrence> = {
      recurrenceType,
      type: recurrenceType,
      interval: Number(interval) || 1,
      daysOfWeek: recurrenceType === RecurrenceType.WEEKLY ? daysOfWeek : undefined,
      dayOfMonth: recurrenceType === RecurrenceType.MONTHLY ? Number(dayOfMonth) || 1 : undefined,
      endType,
      endDate: endType === RecurrenceEndType.ON_DATE && endDate ? new Date(endDate).toISOString() : undefined,
      occurrenceCount: endType === RecurrenceEndType.AFTER_OCCURRENCES ? Number(occurrenceCount) || 1 : undefined,
    };

    if (onChange) {
      onChange(payload);
      setIsOpen(false);
      return;
    }

    if (taskId) {
      try {
        if (isRecurring || serverRecurrence) {
          await updateRecurrence({ taskId, config: payload }).unwrap();
          toast.success('Recurrence pattern updated');
        } else {
          await addRecurrence({ taskId, config: payload }).unwrap();
          toast.success('Task recurrence enabled');
        }
        setIsOpen(false);
      } catch {
        toast.error('Failed to save recurrence configuration');
      }
    }
  };

  const handleRemove = async () => {
    if (onChange) {
      onChange(null);
      setIsOpen(false);
      return;
    }

    if (taskId) {
      try {
        await deleteRecurrence(taskId).unwrap();
        toast.success('Recurrence removed from task');
        setIsOpen(false);
      } catch {
        toast.error('Failed to remove recurrence');
      }
    }
  };

  const isSaving = isAdding || isUpdating || isDeleting;

  // Compute concise label
  const summaryText = React.useMemo(() => {
    const active = isRecurring || !!serverRecurrence || !!initialConfig;
    if (!active) return 'Does not repeat';

    const typeStr = recurrenceType.toLowerCase();
    if (interval === 1) {
      return `Repeats ${typeStr}`;
    }
    return `Repeats every ${interval} ${typeStr === 'daily' ? 'days' : typeStr === 'weekly' ? 'weeks' : 'months'}`;
  }, [isRecurring, serverRecurrence, initialConfig, recurrenceType, interval]);

  return (
    <div className="relative">
      {/* Trigger Button */}
      <button
        type="button"
        disabled={!canEdit}
        onClick={() => setIsOpen((prev) => !prev)}
        className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs border transition-colors cursor-pointer ${
          isRecurring || !!serverRecurrence || !!initialConfig
            ? 'bg-cyan-950/40 border-cyan-800/60 text-cyan-300 hover:bg-cyan-900/40'
            : 'bg-zinc-900/60 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60'
        }`}
        title="Configure Task Recurrence"
      >
        <Repeat className="w-3.5 h-3.5 text-cyan-400" />
        <span className="font-medium">{summaryText}</span>
        <ChevronDown className="w-3 h-3 text-zinc-500 ml-0.5" />
      </button>

      {/* Popover Card */}
      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
            aria-hidden="true"
          />

          <div className="absolute left-0 top-9 z-50 w-72 sm:w-80 rounded-md border border-zinc-800 bg-[#09090b] shadow-2xl p-3.5 space-y-3.5 text-xs text-zinc-300 animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between pb-2 border-b border-zinc-800/80">
              <div className="flex items-center gap-1.5 font-semibold text-zinc-100">
                <Repeat className="w-3.5 h-3.5 text-cyan-400" />
                <span>Repeat Task</span>
              </div>
              <button
                type="button"
                onClick={() => setIsOpen(false)}
                className="text-zinc-500 hover:text-zinc-300 p-0.5"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>

            {/* Recurrence Frequency */}
            <div className="space-y-1">
              <label className="text-[11px] font-medium text-zinc-400">Frequency</label>
              <select
                value={recurrenceType}
                onChange={(e) => setRecurrenceType(e.target.value as RecurrenceType)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded-md px-2.5 py-1.5 text-xs text-zinc-200 focus:outline-hidden focus:border-zinc-700"
              >
                <option value={RecurrenceType.DAILY}>Daily</option>
                <option value={RecurrenceType.WEEKLY}>Weekly</option>
                <option value={RecurrenceType.MONTHLY}>Monthly</option>
                <option value={RecurrenceType.YEARLY}>Yearly</option>
                <option value={RecurrenceType.CUSTOM}>Custom</option>
              </select>
            </div>

            {/* Interval */}
            <div className="space-y-1">
              <label className="text-[11px] font-medium text-zinc-400">
                Repeat every {recurrenceType.toLowerCase().replace('ly', '')}(s)
              </label>
              <Input
                type="number"
                min={1}
                max={99}
                value={interval}
                onChange={(e) => setInterval(Math.max(1, parseInt(e.target.value, 10) || 1))}
                className="h-8 bg-zinc-900 border-zinc-800 text-xs"
              />
            </div>

            {/* Days of Week (for Weekly) */}
            {recurrenceType === RecurrenceType.WEEKLY && (
              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-zinc-400">Repeat on</label>
                <div className="flex items-center justify-between gap-1">
                  {DAYS_OF_WEEK.map((day) => {
                    const isSelected = daysOfWeek.includes(day.value);
                    return (
                      <button
                        key={day.value}
                        type="button"
                        onClick={() => toggleDay(day.value)}
                        className={`h-7 w-8 rounded-sm text-[11px] font-medium transition-colors cursor-pointer border ${
                          isSelected
                            ? 'bg-cyan-600 border-cyan-500 text-white'
                            : 'bg-zinc-900 border-zinc-800 text-zinc-400 hover:text-zinc-200'
                        }`}
                      >
                        {day.label}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Ends Condition */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Ends</label>
              <div className="space-y-1.5">
                <label className="flex items-center gap-2 cursor-pointer text-xs">
                  <input
                    type="radio"
                    name="endType"
                    checked={endType === RecurrenceEndType.NEVER}
                    onChange={() => setEndType(RecurrenceEndType.NEVER)}
                    className="accent-cyan-500"
                  />
                  <span>Never</span>
                </label>

                <div className="flex items-center gap-2 text-xs">
                  <input
                    type="radio"
                    name="endType"
                    checked={endType === RecurrenceEndType.ON_DATE}
                    onChange={() => setEndType(RecurrenceEndType.ON_DATE)}
                    className="accent-cyan-500 shrink-0"
                  />
                  <span className="shrink-0">On date</span>
                  {endType === RecurrenceEndType.ON_DATE && (
                    <Input
                      type="date"
                      value={endDate}
                      onChange={(e) => setEndDate(e.target.value)}
                      className="h-7 bg-zinc-900 border-zinc-800 text-xs py-0 px-2"
                    />
                  )}
                </div>

                <div className="flex items-center gap-2 text-xs">
                  <input
                    type="radio"
                    name="endType"
                    checked={endType === RecurrenceEndType.AFTER_OCCURRENCES}
                    onChange={() => setEndType(RecurrenceEndType.AFTER_OCCURRENCES)}
                    className="accent-cyan-500 shrink-0"
                  />
                  <span className="shrink-0">After</span>
                  {endType === RecurrenceEndType.AFTER_OCCURRENCES && (
                    <div className="flex items-center gap-1">
                      <Input
                        type="number"
                        min={1}
                        max={999}
                        value={occurrenceCount}
                        onChange={(e) => setOccurrenceCount(parseInt(e.target.value, 10) || 1)}
                        className="h-7 w-16 bg-zinc-900 border-zinc-800 text-xs py-0 px-1.5"
                      />
                      <span className="text-[11px] text-zinc-500">occurrences</span>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center justify-between pt-2 border-t border-zinc-800">
              {(isRecurring || serverRecurrence || initialConfig) && (
                <button
                  type="button"
                  onClick={handleRemove}
                  disabled={isSaving}
                  className="flex items-center gap-1 text-[11px] text-rose-400 hover:text-rose-300 disabled:opacity-40 transition-colors"
                >
                  <Trash2 className="w-3 h-3" />
                  <span>Remove</span>
                </button>
              )}

              <div className="flex items-center gap-1.5 ml-auto">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsOpen(false)}
                  className="h-7 text-xs border-zinc-800"
                >
                  Cancel
                </Button>
                <Button
                  type="button"
                  size="sm"
                  onClick={handleSave}
                  disabled={isSaving}
                  className="h-7 text-xs bg-cyan-600 hover:bg-cyan-500 text-white"
                >
                  {isSaving ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : null}
                  Save
                </Button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

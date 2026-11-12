'use client';

import * as React from 'react';
import { useState, useEffect, useRef } from 'react';
import { Clock, Plus, Play, Pause, Square, RotateCcw, History } from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetTimeEntriesQuery,
  useCreateTimeEntryMutation,
} from '@/lib/store/api/timeEntriesApi';
import type { TaskTimeTrackingProps } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export const TaskTimeTracking = React.memo(function TaskTimeTracking({ taskId, estimatedHours, canEdit }: TaskTimeTrackingProps) {
  const { data: entries = [], isLoading } = useGetTimeEntriesQuery(
    { taskId },
    { skip: !taskId },
  );

  const [createTimeEntry, { isLoading: isLogging }] = useCreateTimeEntryMutation();

  const [isLoggingOpen, setIsLoggingOpen] = useState(false);
  const [hours, setHours] = useState('');
  const [minutes, setMinutes] = useState('');
  const [description, setDescription] = useState('');

  // Live Stopwatch State
  const [isTimerRunning, setIsTimerRunning] = useState(false);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);

  useEffect(() => {
    let interval: NodeJS.Timeout | null = null;
    if (isTimerRunning) {
      interval = setInterval(() => {
        setElapsedSeconds((prev) => prev + 1);
      }, 1000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isTimerRunning]);

  const formatStopwatch = (totalSecs: number) => {
    const h = Math.floor(totalSecs / 3600);
    const m = Math.floor((totalSecs % 3600) / 60);
    const s = totalSecs % 60;
    const pad = (n: number) => n.toString().padStart(2, '0');
    return `${pad(h)}:${pad(m)}:${pad(s)}`;
  };

  const handleStartTimer = () => {
    setIsTimerRunning(true);
    toast.info('Stopwatch timer started');
  };

  const handlePauseTimer = () => {
    setIsTimerRunning(false);
  };

  const handleResetTimer = () => {
    setIsTimerRunning(false);
    setElapsedSeconds(0);
  };

  const handleStopAndLog = () => {
    setIsTimerRunning(false);
    const totalMinutes = Math.max(1, Math.round(elapsedSeconds / 60));
    const h = Math.floor(totalMinutes / 60);
    const m = totalMinutes % 60;

    setHours(h > 0 ? h.toString() : '');
    setMinutes(m > 0 ? m.toString() : '1');
    setIsLoggingOpen(true);
    setElapsedSeconds(0);
    toast.success('Timer stopped. Review and save your work log below.');
  };

  // Total seconds calculated from entries
  const totalSeconds = entries.reduce((acc, e) => acc + (e.duration || 0), 0);
  const totalHours = Math.floor(totalSeconds / 3600);
  const totalMins = Math.floor((totalSeconds % 3600) / 60);

  const handleLogTime = async (e: React.FormEvent) => {
    e.preventDefault();
    const h = parseInt(hours, 10) || 0;
    const m = parseInt(minutes, 10) || 0;
    const durationInSeconds = h * 3600 + m * 60;

    if (durationInSeconds <= 0) {
      toast.error('Please enter valid hours or minutes');
      return;
    }

    try {
      await createTimeEntry({
        taskId,
        duration: durationInSeconds,
        description: description.trim() || undefined,
      }).unwrap();

      toast.success('Time logged successfully');
      setHours('');
      setMinutes('');
      setDescription('');
      setIsLoggingOpen(false);
    } catch {
      toast.error('Failed to log time');
    }
  };

  return (
    <div className="space-y-3 pt-4 border-t border-zinc-800">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4 text-zinc-400" />
          <span className="text-xs font-bold text-zinc-300">Time Tracking</span>
        </div>

        {canEdit && (
          <button
            type="button"
            onClick={() => setIsLoggingOpen(!isLoggingOpen)}
            className="text-[11px] text-zinc-400 hover:text-zinc-200 transition-colors flex items-center gap-1 cursor-pointer"
          >
            <Plus className="h-3 w-3" />
            <span>Log Time</span>
          </button>
        )}
      </div>

      {/* Stopwatch Live Timer Bar */}
      {canEdit && (
        <div className="p-2.5 rounded-md bg-zinc-950/80 border border-zinc-800 flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span
              className={`h-2.5 w-2.5 rounded-full ${
                isTimerRunning
                  ? 'bg-emerald-500 animate-pulse'
                  : elapsedSeconds > 0
                  ? 'bg-amber-500'
                  : 'bg-zinc-600'
              }`}
            />
            <span className="font-mono text-sm font-bold text-zinc-100 tracking-wider">
              {formatStopwatch(elapsedSeconds)}
            </span>
          </div>

          <div className="flex items-center gap-1.5">
            {!isTimerRunning && elapsedSeconds === 0 ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleStartTimer}
                className="h-7 px-2.5 text-[11px] gap-1 border-emerald-900/60 bg-emerald-950/30 text-emerald-400 hover:bg-emerald-950/60 cursor-pointer"
              >
                <Play className="h-3 w-3 fill-emerald-400" />
                <span>Start</span>
              </Button>
            ) : isTimerRunning ? (
              <>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handlePauseTimer}
                  className="h-7 px-2 text-[11px] gap-1 border-amber-900/60 bg-amber-950/30 text-amber-400 hover:bg-amber-950/60 cursor-pointer"
                >
                  <Pause className="h-3 w-3" />
                  <span>Pause</span>
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  size="sm"
                  onClick={handleStopAndLog}
                  className="h-7 px-2 text-[11px] gap-1 cursor-pointer"
                >
                  <Square className="h-3 w-3 fill-current" />
                  <span>Stop & Log</span>
                </Button>
              </>
            ) : (
              <>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleStartTimer}
                  className="h-7 px-2 text-[11px] gap-1 border-emerald-900/60 text-emerald-400 cursor-pointer"
                >
                  <Play className="h-3 w-3 fill-emerald-400" />
                  <span>Resume</span>
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  size="sm"
                  onClick={handleStopAndLog}
                  className="h-7 px-2 text-[11px] gap-1 cursor-pointer"
                >
                  <Square className="h-3 w-3 fill-current" />
                  <span>Log</span>
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleResetTimer}
                  className="h-7 w-7 p-0 text-zinc-500 hover:text-zinc-300 cursor-pointer"
                  title="Reset Timer"
                >
                  <RotateCcw className="h-3 w-3" />
                </Button>
              </>
            )}
          </div>
        </div>
      )}

      {/* Summary Cards */}
      <div className="grid grid-cols-2 gap-3 p-3 rounded-md bg-zinc-950/60 border border-zinc-850 text-xs">
        <div>
          <span className="text-[10px] uppercase font-mono text-zinc-500">Logged Time</span>
          <div className="text-sm font-bold text-zinc-100 font-mono mt-0.5">
            {totalHours}h {totalMins}m
          </div>
        </div>

        <div>
          <span className="text-[10px] uppercase font-mono text-zinc-500">Estimated</span>
          <div className="text-sm font-bold text-zinc-400 font-mono mt-0.5">
            {estimatedHours ? `${estimatedHours}h` : 'No estimate'}
          </div>
        </div>
      </div>

      {/* Log Time Form */}
      {isLoggingOpen && (
        <form onSubmit={handleLogTime} className="p-3 rounded-md bg-zinc-950/80 border border-zinc-800 space-y-3">
          <span className="text-xs font-semibold text-zinc-200 block">Record Spent Time</span>

          <div className="grid grid-cols-2 gap-2">
            <div>
              <span className="text-[10px] text-zinc-400">Hours</span>
              <Input
                type="number"
                min="0"
                value={hours}
                onChange={(e) => setHours(e.target.value)}
                placeholder="1"
                className="bg-zinc-900 border-zinc-800 text-xs text-zinc-100 h-7"
              />
            </div>
            <div>
              <span className="text-[10px] text-zinc-400">Minutes</span>
              <Input
                type="number"
                min="0"
                max="59"
                value={minutes}
                onChange={(e) => setMinutes(e.target.value)}
                placeholder="30"
                className="bg-zinc-900 border-zinc-800 text-xs text-zinc-100 h-7"
              />
            </div>
          </div>

          <Input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Work notes (e.g. Unit testing & debugging)..."
            className="bg-zinc-900 border-zinc-800 text-xs text-zinc-100 h-7"
          />

          <div className="flex gap-2">
            <Button
              type="submit"
              size="sm"
              variant="primary"
              isLoading={isLogging}
              className="text-xs h-7"
            >
              Save Time Log
            </Button>
            <Button
              type="button"
              size="sm"
              variant="secondary"
              onClick={() => setIsLoggingOpen(false)}
              className="text-xs h-7"
            >
              Cancel
            </Button>
          </div>
        </form>
      )}

      {/* Time Entries History */}
      {entries.length > 0 && (
        <div className="space-y-1 pt-1">
          <span className="text-[10px] uppercase font-mono text-zinc-500">Recent Logs</span>
          <div className="space-y-1 max-h-32 overflow-y-auto">
            {entries.slice(0, 5).map((entry) => {
              const h = Math.floor(entry.duration / 3600);
              const m = Math.floor((entry.duration % 3600) / 60);

              return (
                <div
                  key={entry.id}
                  className="flex items-center justify-between p-2 rounded bg-zinc-950/40 text-[11px] text-zinc-400 border border-zinc-850"
                >
                  <div className="flex items-center gap-2 truncate">
                    <History className="h-3 w-3 text-zinc-600 shrink-0" />
                    <span className="truncate">{entry.description || 'Logged work'}</span>
                  </div>
                  <span className="font-mono text-zinc-200 shrink-0 font-medium">
                    {h > 0 ? `${h}h ` : ''}{m}m
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
});


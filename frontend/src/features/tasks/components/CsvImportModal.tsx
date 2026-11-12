'use client';

import React, { useState, useRef, useMemo } from 'react';
import {
  Upload,
  FileSpreadsheet,
  AlertCircle,
  CheckCircle2,
  X,
  Download,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  useBulkCreateTasksMutation,
  useGetTaskStatusesByProjectQuery,
} from '@/lib/store/api/tasksApi';
import type { BulkTaskItem } from '@/types';

interface CsvImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  sprintId?: string;
}

interface ParsedRow {
  index: number;
  data: BulkTaskItem;
  isValid: boolean;
  error?: string;
}

export const CsvImportModal: React.FC<CsvImportModalProps> = ({
  isOpen,
  onClose,
  projectId,
  sprintId,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFileName, setSelectedFileName] = useState<string | null>(null);
  const [parsedRows, setParsedRows] = useState<ParsedRow[]>([]);
  const [selectedStatusId, setSelectedStatusId] = useState<string>('');
  const [isDragging, setIsDragging] = useState(false);

  // Fetch project statuses
  const { data: statuses = [], isLoading: isStatusesLoading } =
    useGetTaskStatusesByProjectQuery(projectId, { skip: !projectId || !isOpen });

  // Bulk create mutation
  const [bulkCreateTasks, { isLoading: isImporting }] = useBulkCreateTasksMutation();

  // Set default status when statuses load
  React.useEffect(() => {
    if (statuses.length > 0 && !selectedStatusId) {
      const defaultStatus =
        statuses.find((s) => s.category === 'TODO') || statuses[0];
      if (defaultStatus) {
        setSelectedStatusId(defaultStatus.id);
      }
    }
  }, [statuses, selectedStatusId]);

  // Reset state on close
  const handleClose = () => {
    setSelectedFileName(null);
    setParsedRows([]);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
    onClose();
  };

  // Robust CSV Line Parser
  const parseCSVLine = (line: string): string[] => {
    const values: string[] = [];
    let current = '';
    let insideQuotes = false;

    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      if (char === '"') {
        if (insideQuotes && line[i + 1] === '"') {
          current += '"';
          i++; // Skip escaped quote
        } else {
          insideQuotes = !insideQuotes;
        }
      } else if (char === ',' && !insideQuotes) {
        values.push(current.trim());
        current = '';
      } else {
        current += char;
      }
    }
    values.push(current.trim());
    return values;
  };

  // Normalize priorities
  const normalizePriority = (raw?: string): 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT' => {
    if (!raw) return 'MEDIUM';
    const clean = raw.trim().toUpperCase();
    if (clean === 'URGENT') return 'URGENT';
    if (clean === 'HIGH') return 'HIGH';
    if (clean === 'LOW') return 'LOW';
    return 'MEDIUM';
  };

  // Normalize types
  const normalizeType = (raw?: string): 'TASK' | 'BUG' | 'STORY' | 'EPIC' => {
    if (!raw) return 'TASK';
    const clean = raw.trim().toUpperCase();
    if (clean === 'BUG') return 'BUG';
    if (clean === 'STORY') return 'STORY';
    if (clean === 'EPIC') return 'EPIC';
    return 'TASK';
  };

  // Normalize due date
  const normalizeDueDate = (raw?: string): string | undefined => {
    if (!raw) return undefined;
    const timestamp = Date.parse(raw);
    if (isNaN(timestamp)) return undefined;
    return new Date(timestamp).toISOString();
  };

  // Process CSV file content
  const processCSVContent = (content: string) => {
    const lines = content.split(/\r?\n/).filter((l) => l.trim().length > 0);
    if (lines.length < 2) {
      toast.error('CSV file must have a header row and at least one task row');
      return;
    }

    const headers = parseCSVLine(lines[0]).map((h) => h.toLowerCase().replace(/[\s_-]+/g, ''));

    // Map column indexes
    const titleIdx = headers.findIndex((h) => h.includes('title') || h.includes('task') || h.includes('name'));
    const descIdx = headers.findIndex((h) => h.includes('desc') || h.includes('summary'));
    const priorityIdx = headers.findIndex((h) => h.includes('prio'));
    const typeIdx = headers.findIndex((h) => h.includes('type') || h.includes('kind'));
    const dateIdx = headers.findIndex((h) => h.includes('due') || h.includes('deadline') || h.includes('date'));

    if (titleIdx === -1) {
      toast.error('Could not find a "Title" or "Task" column header in the CSV');
      return;
    }

    const rows: ParsedRow[] = [];

    for (let i = 1; i < lines.length; i++) {
      const values = parseCSVLine(lines[i]);
      const rawTitle = values[titleIdx] || '';
      const title = rawTitle.replace(/^"|"$/g, '').trim();

      const description = descIdx !== -1 && values[descIdx] ? values[descIdx].replace(/^"|"$/g, '').trim() : undefined;
      const priority = priorityIdx !== -1 ? normalizePriority(values[priorityIdx]) : 'MEDIUM';
      const type = typeIdx !== -1 ? normalizeType(values[typeIdx]) : 'TASK';
      const dueDate = dateIdx !== -1 ? normalizeDueDate(values[dateIdx]) : undefined;

      const isValid = title.length > 0;
      rows.push({
        index: i,
        data: {
          title: title || 'Untitled Task',
          description,
          priority,
          type,
          dueDate,
        },
        isValid,
        error: isValid ? undefined : 'Title is required',
      });
    }

    setParsedRows(rows);
    toast.info(`Parsed ${rows.length} rows from CSV`);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.csv') && file.type !== 'text/csv') {
      toast.error('Please upload a valid .csv file');
      return;
    }

    setSelectedFileName(file.name);
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        processCSVContent(text);
      }
    };
    reader.readAsText(file);
  };

  // Drag and drop handlers
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.csv') && file.type !== 'text/csv') {
      toast.error('Please upload a valid .csv file');
      return;
    }

    setSelectedFileName(file.name);
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        processCSVContent(text);
      }
    };
    reader.readAsText(file);
  };

  // Sample CSV template download
  const handleDownloadSample = () => {
    const sample = `Title,Description,Priority,Type,Due Date\n"Fix login authentication issue","Users report 401 on token expiration","HIGH","BUG","2026-11-20"\n"Design analytics dashboard wireframes","High fidelity wireframes for KPIs","MEDIUM","TASK","2026-11-25"\n"Deploy Redis cache cluster","Upgrade memory limit to 4GB","LOW","TASK","2026-12-01"`;
    const blob = new Blob([sample], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', 'orbito_tasks_template.csv');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // Valid rows count
  const validTasks = useMemo(() => {
    return parsedRows.filter((r) => r.isValid).map((r) => r.data);
  }, [parsedRows]);

  // Submission
  const handleImport = async () => {
    if (!selectedStatusId) {
      toast.error('Please select a target status for imported tasks');
      return;
    }

    if (validTasks.length === 0) {
      toast.error('No valid tasks to import');
      return;
    }

    try {
      const res = await bulkCreateTasks({
        projectId,
        statusId: selectedStatusId,
        sprintId: sprintId || undefined,
        tasks: validTasks,
      }).unwrap();

      toast.success(`Successfully imported ${res.created} tasks!`);
      if (res.failed > 0) {
        toast.warning(`${res.failed} tasks could not be imported.`);
      }
      handleClose();
    } catch {
      toast.error('Failed to import tasks. Please verify your CSV format.');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Import Tasks from CSV"
      description="Upload a CSV file to bulk create tasks in this project."
    >
      <div className="space-y-4 pt-2">
        {/* File Dropzone */}
        {!selectedFileName ? (
          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border-2 border-dashed rounded-md p-6 flex flex-col items-center justify-center cursor-pointer transition text-center ${
              isDragging
                ? 'border-zinc-400 bg-zinc-900/60'
                : 'border-zinc-800 bg-zinc-950/60 hover:border-zinc-700 hover:bg-zinc-900/40'
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              onChange={handleFileChange}
              className="hidden"
            />
            <Upload className="h-8 w-8 text-zinc-400 mb-2" />
            <span className="text-xs font-medium text-zinc-200">
              Drag & drop your CSV file here, or click to browse
            </span>
            <span className="text-[11px] text-zinc-500 mt-1">
              Supports columns: Title, Description, Priority, Type, Due Date
            </span>

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                handleDownloadSample();
              }}
              className="mt-4 gap-1.5 text-[11px] h-7 border-zinc-700 bg-zinc-900 text-zinc-300"
            >
              <Download className="h-3 w-3" />
              Download Sample CSV
            </Button>
          </div>
        ) : (
          <div className="border border-zinc-800 rounded-md p-3 bg-zinc-900/50 flex items-center justify-between">
            <div className="flex items-center gap-2.5 truncate">
              <FileSpreadsheet className="h-5 w-5 text-emerald-400 shrink-0" />
              <div className="truncate">
                <span className="text-xs font-medium text-zinc-200 block truncate">
                  {selectedFileName}
                </span>
                <span className="text-[10px] text-zinc-400">
                  {parsedRows.length} total rows parsed ({validTasks.length} valid)
                </span>
              </div>
            </div>

            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                setSelectedFileName(null);
                setParsedRows([]);
                if (fileInputRef.current) fileInputRef.current.value = '';
              }}
              className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-200"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        )}

        {/* Target Status Selector */}
        {parsedRows.length > 0 && (
          <div className="space-y-1.5">
            <Label className="text-xs text-zinc-300">Target Status</Label>
            <select
              value={selectedStatusId}
              onChange={(e) => setSelectedStatusId(e.target.value)}
              disabled={isStatusesLoading}
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 rounded px-3 py-1.5 text-xs outline-none focus:border-zinc-700 h-8"
            >
              {statuses.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.category})
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Parsed Preview Table */}
        {parsedRows.length > 0 && (
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-[11px] text-zinc-400">
              <span>Preview (First 10 rows)</span>
              <span>
                {validTasks.length} of {parsedRows.length} rows valid
              </span>
            </div>

            <div className="max-h-56 overflow-y-auto border border-zinc-800 rounded-md bg-zinc-950/80 divide-y divide-zinc-850">
              {parsedRows.slice(0, 10).map((row) => (
                <div
                  key={row.index}
                  className="px-3 py-2 flex items-center justify-between text-xs hover:bg-zinc-900/30 transition"
                >
                  <div className="flex items-center gap-2 truncate pr-2">
                    {row.isValid ? (
                      <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                    ) : (
                      <AlertCircle className="h-3.5 w-3.5 text-red-400 shrink-0" />
                    )}
                    <span
                      className={`truncate font-medium text-[11px] ${
                        row.isValid ? 'text-zinc-200' : 'text-red-300'
                      }`}
                    >
                      {row.data.title}
                    </span>
                  </div>

                  <div className="flex items-center gap-1.5 shrink-0">
                    <Badge variant="outline" className="text-[9px] px-1 py-0">
                      {row.data.priority}
                    </Badge>
                    <Badge variant="outline" className="text-[9px] px-1 py-0">
                      {row.data.type}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleClose}
            disabled={isImporting}
            className="text-xs h-8 text-zinc-400 hover:text-zinc-200"
          >
            Cancel
          </Button>

          <Button
            type="button"
            variant="primary"
            size="sm"
            onClick={handleImport}
            disabled={isImporting || validTasks.length === 0 || !selectedStatusId}
            className="text-xs h-8 gap-1.5"
          >
            {isImporting ? (
              <>
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                <span>Importing...</span>
              </>
            ) : (
              <>
                <Upload className="h-3.5 w-3.5" />
                <span>Import {validTasks.length} Tasks</span>
              </>
            )}
          </Button>
        </div>
      </div>
    </Modal>
  );
};
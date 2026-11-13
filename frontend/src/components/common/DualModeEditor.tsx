'use client';

import * as React from 'react';
import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import {
  Bold,
  Italic,
  Heading,
  List,
  ListOrdered,
  Quote,
  Code,
  Link as LinkIcon,
  AtSign,
  Eye,
  Edit3,
  User as UserIcon,
} from 'lucide-react';
import { useGetWorkspaceMembersQuery } from '@/lib/store/api/workspacesApi';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export interface DualModeEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  minRows?: number;
  disabled?: boolean;
  workspaceId?: string;
  className?: string;
  onBlur?: () => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
}

export function DualModeEditor({
  value,
  onChange,
  placeholder = 'Write in Markdown or use the toolbar... Type @ to mention teammates',
  minRows = 4,
  disabled = false,
  workspaceId,
  className = '',
  onBlur,
  onKeyDown,
}: DualModeEditorProps) {
  const [activeTab, setActiveTab] = useState<'write' | 'preview'>('write');
  const [mentionQuery, setMentionQuery] = useState<string | null>(null);
  const [mentionIndex, setMentionIndex] = useState<number>(0);
  const [cursorPosition, setCursorPosition] = useState<number>(0);

  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const mentionMenuRef = useRef<HTMLDivElement | null>(null);

  // Load workspace members for @mention autocomplete
  const { data: members = [] } = useGetWorkspaceMembersQuery(
    workspaceId || '',
    { skip: !workspaceId }
  );

  // Filter members based on mentionQuery
  const filteredMembers = useMemo(() => {
    if (mentionQuery === null) return [];
    const query = mentionQuery.toLowerCase();
    return members
      .filter((m) => {
        const first = m.user?.firstName?.toLowerCase() || '';
        const last = m.user?.lastName?.toLowerCase() || '';
        const email = m.user?.email?.toLowerCase() || '';
        const fullName = `${first} ${last}`.trim();
        return (
          fullName.includes(query) ||
          first.includes(query) ||
          last.includes(query) ||
          email.includes(query)
        );
      })
      .slice(0, 5); // top 5 results
  }, [members, mentionQuery]);

  // Insert formatting into textarea
  const insertFormat = useCallback(
    (prefix: string, suffix: string = prefix, defaultPlaceholder: string = '') => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const selected = value.substring(start, end);
      const textToInsert = selected || defaultPlaceholder;

      const updated =
        value.substring(0, start) + prefix + textToInsert + suffix + value.substring(end);
      onChange(updated);

      requestAnimationFrame(() => {
        textarea.focus();
        const nextPos = selected
          ? end + prefix.length + suffix.length
          : start + prefix.length + textToInsert.length;
        textarea.setSelectionRange(nextPos, nextPos);
      });
    },
    [value, onChange]
  );

  // Insert block formatting at start of line
  const insertBlockPrefix = useCallback(
    (prefix: string) => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      const start = textarea.selectionStart;
      const lastNewline = value.lastIndexOf('\n', start - 1);
      const lineStart = lastNewline === -1 ? 0 : lastNewline + 1;

      const updated =
        value.substring(0, lineStart) + prefix + value.substring(lineStart);
      onChange(updated);

      requestAnimationFrame(() => {
        textarea.focus();
        const nextPos = start + prefix.length;
        textarea.setSelectionRange(nextPos, nextPos);
      });
    },
    [value, onChange]
  );

  // Check for @mention trigger
  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const nextVal = e.target.value;
    onChange(nextVal);

    const pos = e.target.selectionStart;
    setCursorPosition(pos);

    const textBeforeCursor = nextVal.substring(0, pos);
    const match = textBeforeCursor.match(/(?:^|\s)@([a-zA-Z0-9._]*)$/);

    if (match) {
      setMentionQuery(match[1]);
      setMentionIndex(0);
    } else {
      setMentionQuery(null);
    }
  };

  // Complete mention
  const handleSelectMention = useCallback(
    (memberName: string) => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      const textBeforeCursor = value.substring(0, cursorPosition);
      const textAfterCursor = value.substring(cursorPosition);
      const match = textBeforeCursor.match(/(?:^|\s)@([a-zA-Z0-9._]*)$/);

      if (match) {
        const matchStart = textBeforeCursor.lastIndexOf(`@${match[1]}`);
        const mentionText = `@[${memberName}] `;
        const updated =
          value.substring(0, matchStart) + mentionText + textAfterCursor;
        onChange(updated);

        const nextPos = matchStart + mentionText.length;
        requestAnimationFrame(() => {
          textarea.focus();
          textarea.setSelectionRange(nextPos, nextPos);
        });
      }

      setMentionQuery(null);
    },
    [value, cursorPosition, onChange]
  );

  // Keyboard navigation for mentions
  const handleTextareaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (mentionQuery !== null && filteredMembers.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setMentionIndex((prev) => (prev + 1) % filteredMembers.length);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setMentionIndex((prev) =>
          prev === 0 ? filteredMembers.length - 1 : prev - 1
        );
        return;
      }
      if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault();
        const selected = filteredMembers[mentionIndex];
        if (selected) {
          const name = `${selected.user?.firstName || ''} ${selected.user?.lastName || ''}`.trim() || selected.user?.email || 'Member';
          handleSelectMention(name);
        }
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        setMentionQuery(null);
        return;
      }
    }

    if (onKeyDown) {
      onKeyDown(e);
    }
  };

  // Pre-process markdown to render @[Name] with custom styling
  const formattedPreview = useMemo(() => {
    if (!value) return '';
    // Replace @[Name] with styled HTML-safe markdown badge
    return value.replace(/@\[([^\]]+)\]/g, '**`@$1`**');
  }, [value]);

  return (
    <div
      className={`border border-zinc-800 rounded-md bg-zinc-950/80 overflow-hidden focus-within:border-zinc-700 transition-colors ${className}`}
    >
      {/* Top Toolbar */}
      <div className="flex items-center justify-between border-b border-zinc-800/80 bg-zinc-900/60 px-2 py-1 select-none">
        {/* Formatting Actions (Active only in Write mode) */}
        <div className="flex items-center gap-0.5">
          {activeTab === 'write' ? (
            <>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertFormat('**', '**', 'bold text')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Bold (Ctrl+B)"
              >
                <Bold className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertFormat('*', '*', 'italic text')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Italic (Ctrl+I)"
              >
                <Italic className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertBlockPrefix('### ')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Heading 3"
              >
                <Heading className="w-3.5 h-3.5" />
              </button>
              <div className="h-3 w-px bg-zinc-800 mx-1" />
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertBlockPrefix('- ')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Bullet List"
              >
                <List className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertBlockPrefix('1. ')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Numbered List"
              >
                <ListOrdered className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertBlockPrefix('> ')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Quote"
              >
                <Quote className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertFormat('```\n', '\n```', 'code')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Code Block"
              >
                <Code className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => insertFormat('[', '](https://)', 'link title')}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors"
                title="Insert Link"
              >
                <LinkIcon className="w-3.5 h-3.5" />
              </button>
              <button
                type="button"
                disabled={disabled}
                onClick={() => {
                  insertFormat('@', '', '');
                  setMentionQuery('');
                }}
                className="p-1 rounded-sm text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-40 transition-colors cursor-pointer"
                title="Mention Collaborator (@)"
              >
                <AtSign className="w-3.5 h-3.5" />
              </button>
            </>
          ) : (
            <span className="text-[11px] font-medium text-zinc-400 px-1">
              Markdown Live Preview
            </span>
          )}
        </div>

        {/* Mode Switcher Tabs */}
        <div className="flex items-center gap-1 bg-zinc-950 p-0.5 rounded-sm border border-zinc-800">
          <button
            type="button"
            onClick={() => setActiveTab('write')}
            className={`flex items-center gap-1 px-2 py-0.5 rounded-xs text-[11px] font-medium transition-colors ${
              activeTab === 'write'
                ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Edit3 className="w-3 h-3" />
            <span>Write</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('preview')}
            className={`flex items-center gap-1 px-2 py-0.5 rounded-xs text-[11px] font-medium transition-colors ${
              activeTab === 'preview'
                ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Eye className="w-3 h-3" />
            <span>Preview</span>
          </button>
        </div>
      </div>

      {/* Editor Body */}
      <div className="relative">
        {activeTab === 'write' ? (
          <textarea
            ref={textareaRef}
            rows={minRows}
            value={value}
            disabled={disabled}
            onChange={handleInputChange}
            onKeyDown={handleTextareaKeyDown}
            onBlur={onBlur}
            placeholder={placeholder}
            className="w-full bg-transparent border-0 resize-y p-3 text-xs text-zinc-100 placeholder:text-zinc-500 focus:outline-hidden font-mono leading-relaxed selection:bg-zinc-800"
          />
        ) : (
          <div className="p-3 min-h-[100px] text-xs leading-relaxed text-zinc-300 prose prose-invert prose-xs max-w-none [&_p]:leading-relaxed [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:bg-zinc-900 [&_code]:border [&_code]:border-zinc-800 [&_code]:rounded-xs [&_code]:text-violet-300 [&_pre]:bg-zinc-950 [&_pre]:p-2.5 [&_pre]:rounded-md [&_pre]:border [&_pre]:border-zinc-800">
            {value.trim() ? (
              <ReactMarkdown>{formattedPreview}</ReactMarkdown>
            ) : (
              <p className="text-zinc-500 italic">Nothing to preview</p>
            )}
          </div>
        )}

        {/* @Mention Autocomplete Dropdown */}
        {mentionQuery !== null && filteredMembers.length > 0 && activeTab === 'write' && (
          <div
            ref={mentionMenuRef}
            className="absolute left-3 top-10 z-50 w-64 rounded-md border border-zinc-800 bg-[#09090b] shadow-xl p-1 space-y-0.5 max-h-48 overflow-y-auto"
          >
            <div className="px-2 py-1 text-[10px] uppercase font-bold text-zinc-500 border-b border-zinc-800/60">
              Mention Member
            </div>
            {filteredMembers.map((member, idx) => {
              const name = `${member.user?.firstName || ''} ${member.user?.lastName || ''}`.trim() || member.user?.email || 'Member';
              const isSelected = idx === mentionIndex;

              return (
                <button
                  key={member.id}
                  type="button"
                  onClick={() => handleSelectMention(name)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 rounded-sm text-left text-xs cursor-pointer transition-colors ${
                    isSelected
                      ? 'bg-zinc-800 text-zinc-100 border border-zinc-700'
                      : 'text-zinc-300 hover:bg-zinc-900'
                  }`}
                >
                  <Avatar className="h-5 w-5 rounded-full border border-zinc-800">
                    <AvatarFallback className="text-[10px] font-bold bg-zinc-800 text-zinc-300">
                      {name[0] || 'U'}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 truncate">
                    <div className="truncate font-medium">{name}</div>
                    <div className="text-[10px] text-zinc-500 truncate">
                      {member.user?.email}
                    </div>
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

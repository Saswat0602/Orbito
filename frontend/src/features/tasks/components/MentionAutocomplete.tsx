'use client';

import * as React from 'react';
import { useState, useEffect, useRef } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

export interface MentionMember {
  id: string;
  name: string;
  email?: string;
  avatar?: string | null;
  role?: string;
}

interface MentionAutocompleteProps {
  isOpen: boolean;
  searchQuery: string;
  members: MentionMember[];
  onSelect: (member: MentionMember) => void;
  onClose: () => void;
  anchorRef?: React.RefObject<HTMLElement | null>;
  className?: string;
}

export function MentionAutocomplete({
  isOpen,
  searchQuery,
  members,
  onSelect,
  onClose,
  className = '',
}: MentionAutocompleteProps) {
  const [selectedIndex, setSelectedIndex] = useState(0);

  const filteredMembers = React.useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return members.slice(0, 5);
    return members
      .filter(
        (m) =>
          m.name.toLowerCase().includes(query) ||
          (m.email && m.email.toLowerCase().includes(query))
      )
      .slice(0, 5);
  }, [members, searchQuery]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [searchQuery]);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (filteredMembers.length === 0) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev + 1) % filteredMembers.length);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev - 1 + filteredMembers.length) % filteredMembers.length);
      } else if (e.key === 'Enter' || e.key === 'Tab') {
        if (filteredMembers[selectedIndex]) {
          e.preventDefault();
          onSelect(filteredMembers[selectedIndex]);
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown, true);
    return () => window.removeEventListener('keydown', handleKeyDown, true);
  }, [isOpen, filteredMembers, selectedIndex, onSelect, onClose]);

  if (!isOpen) return null;

  return (
    <div
      className={`absolute z-50 w-64 rounded-md border border-zinc-800 bg-zinc-900/95 shadow-2xl backdrop-blur-md p-1 overflow-hidden transition-all ${className}`}
      style={{ bottom: '100%', marginBottom: '4px', left: '0' }}
    >
      <div className="px-2 py-1 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider border-b border-zinc-800/80 mb-1">
        Mention Team Member
      </div>

      {filteredMembers.length === 0 ? (
        <div className="py-2.5 px-3 text-center text-xs text-zinc-500">
          No matching members found
        </div>
      ) : (
        <div className="space-y-0.5">
          {filteredMembers.map((member, idx) => {
            const isSelected = idx === selectedIndex;
            const initials = member.name
              .split(' ')
              .map((n) => n[0])
              .join('')
              .toUpperCase()
              .slice(0, 2) || 'U';

            return (
              <button
                key={member.id}
                type="button"
                onMouseDown={(e) => {
                  e.preventDefault(); // Don't steal input focus
                  onSelect(member);
                }}
                className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-sm text-xs transition-colors cursor-pointer text-left ${
                  isSelected
                    ? 'bg-zinc-800 text-zinc-100 font-medium'
                    : 'text-zinc-300 hover:bg-zinc-850 hover:text-zinc-200'
                }`}
              >
                <div className="flex items-center gap-2 min-w-0">
                  <Avatar className="h-5 w-5 rounded-sm border border-zinc-700 shrink-0">
                    {member.avatar && <AvatarImage src={member.avatar} alt={member.name} />}
                    <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300 font-bold">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                  <span className="truncate text-xs">{member.name}</span>
                </div>

                {member.role && (
                  <span className="text-[10px] font-mono text-zinc-500 shrink-0">
                    {member.role}
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

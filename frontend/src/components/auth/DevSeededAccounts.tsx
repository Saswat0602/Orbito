'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import {
  Sparkles,
  Zap,
  Search,
  Copy,
  Check,
  Shield,
  Briefcase,
  Code,
  Eye,
  Key,
  AlertCircle,
  Loader2,
  ChevronDown,
} from 'lucide-react';
import { toast } from 'sonner';

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  SEEDED_ACCOUNTS,
  DEFAULT_SEEDED_PASSWORD,
  type SeededAccount,
} from './seededUsers';
import { Role } from '@orbito/types';
import { cn } from '@/lib/utils';

interface DevSeededAccountsProps {
  onSelectAccount: (account: SeededAccount, autoSubmit?: boolean) => void;
  selectedEmail?: string;
  isLoggingIn?: boolean;
  className?: string;
}

export function DevSeededAccounts({
  onSelectAccount,
  selectedEmail,
  isLoggingIn = false,
  className,
}: DevSeededAccountsProps) {
  // Closed initially so the login form design stays clean and compact
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedRole, setSelectedRole] = useState<string>('ALL');
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [activeSubmittingEmail, setActiveSubmittingEmail] = useState<string | null>(null);

  const handleCopy = (text: string, key: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    toast.success(`Copied ${label} to clipboard`);
    setTimeout(() => {
      setCopiedKey((curr) => (curr === key ? null : curr));
    }, 2000);
  };

  const handleAutoFill = (account: SeededAccount) => {
    onSelectAccount(account, false);
    toast.info(`Filled credentials for ${account.name}`);
  };

  const handleOneClickLogin = (account: SeededAccount) => {
    setActiveSubmittingEmail(account.email);
    onSelectAccount(account, true);
  };

  const filteredAccounts = useMemo(() => {
    return SEEDED_ACCOUNTS.filter((acc) => {
      const matchesRole =
        selectedRole === 'ALL' || acc.role === selectedRole;
      const term = searchTerm.toLowerCase().trim();
      const matchesSearch =
        !term ||
        acc.name.toLowerCase().includes(term) ||
        acc.email.toLowerCase().includes(term) ||
        acc.username.toLowerCase().includes(term) ||
        acc.title.toLowerCase().includes(term) ||
        acc.role.toLowerCase().includes(term);

      return matchesRole && matchesSearch;
    });
  }, [searchTerm, selectedRole]);

  const getRoleBadge = (role: Role) => {
    switch (role) {
      case Role.SUPER_ADMIN:
        return (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.2 text-[9px] font-semibold uppercase bg-amber-500/10 text-amber-300 border border-amber-500/25 shrink-0">
            <Shield className="h-2.5 w-2.5" />
            Admin
          </span>
        );
      case Role.MANAGER:
        return (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.2 text-[9px] font-semibold uppercase bg-purple-500/10 text-purple-300 border border-purple-500/25 shrink-0">
            <Briefcase className="h-2.5 w-2.5" />
            Manager
          </span>
        );
      case Role.MEMBER:
        return (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.2 text-[9px] font-semibold uppercase bg-emerald-500/10 text-emerald-300 border border-emerald-500/25 shrink-0">
            <Code className="h-2.5 w-2.5" />
            Member
          </span>
        );
      case Role.VIEWER:
        return (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.2 text-[9px] font-semibold uppercase bg-sky-500/10 text-sky-300 border border-sky-500/25 shrink-0">
            <Eye className="h-2.5 w-2.5" />
            Viewer
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.2 text-[9px] font-semibold uppercase bg-zinc-800 text-zinc-300 border border-zinc-700 shrink-0">
            {role}
          </span>
        );
    }
  };

  const getInitials = (name: string) => {
    const parts = name.split(' ');
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  };

  return (
    <div className={cn('pt-2 space-y-2', className)}>
      {/* Sleek Toggle Button that keeps the login card compact and neat initially */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'w-full flex items-center justify-between px-3 py-2 rounded-lg border text-xs transition-all cursor-pointer select-none group',
          isOpen
            ? 'border-zinc-700 bg-zinc-800/60 text-zinc-100 shadow-xs'
            : 'border-zinc-800/80 bg-zinc-950/40 hover:bg-zinc-800/40 text-zinc-400 hover:text-zinc-200'
        )}
        aria-expanded={isOpen}
      >
        <div className="flex items-center gap-2">
          <Sparkles className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
          <span className="font-medium text-zinc-300 group-hover:text-zinc-100 transition-colors">
            Dev Seed Accounts
          </span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-400 font-mono">
            8 users
          </span>
        </div>
        <div className="flex items-center gap-1.5 text-zinc-500 group-hover:text-zinc-300 transition-colors">
          <span className="text-[11px] font-medium">
            {isOpen ? 'Collapse' : 'Quick Login'}
          </span>
          <ChevronDown
            className={cn(
              'h-3.5 w-3.5 transition-transform duration-200',
              isOpen && 'rotate-180'
            )}
          />
        </div>
      </button>

      {/* Expanded Content Dropdown */}
      {isOpen && (
        <div className="space-y-2.5 rounded-lg border border-zinc-800/80 bg-zinc-950/80 p-3 animate-in fade-in-0 duration-150">
          {/* Top bar with password copy */}
          <div className="flex items-center justify-between text-xs">
            <span className="text-[11px] text-zinc-400">
              Click an account to auto-fill or log in:
            </span>
            <button
              type="button"
              onClick={() =>
                handleCopy(DEFAULT_SEEDED_PASSWORD, 'global-pwd', 'password')
              }
              className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-zinc-900 border border-zinc-800 hover:border-zinc-700 text-zinc-300 text-[10px] font-mono transition-colors cursor-pointer"
              title="Click to copy default seed password"
            >
              <Key className="h-2.5 w-2.5 text-amber-400" />
              <span>password123</span>
              {copiedKey === 'global-pwd' ? (
                <Check className="h-2.5 w-2.5 text-emerald-400" />
              ) : (
                <Copy className="h-2.5 w-2.5 text-zinc-500" />
              )}
            </button>
          </div>

          {/* Role Filters */}
          <div className="flex items-center gap-1 overflow-x-auto pb-0.5 text-xs scrollbar-none">
            {[
              { id: 'ALL', label: 'All', count: SEEDED_ACCOUNTS.length },
              {
                id: Role.SUPER_ADMIN,
                label: 'Admin',
                count: SEEDED_ACCOUNTS.filter((a) => a.role === Role.SUPER_ADMIN).length,
              },
              {
                id: Role.MANAGER,
                label: 'Manager',
                count: SEEDED_ACCOUNTS.filter((a) => a.role === Role.MANAGER).length,
              },
              {
                id: Role.MEMBER,
                label: 'Member',
                count: SEEDED_ACCOUNTS.filter((a) => a.role === Role.MEMBER).length,
              },
              {
                id: Role.VIEWER,
                label: 'Viewer',
                count: SEEDED_ACCOUNTS.filter((a) => a.role === Role.VIEWER).length,
              },
            ].map((tab) => {
              const active = selectedRole === tab.id;
              return (
                <button
                  key={tab.id}
                  type="button"
                  onClick={() => setSelectedRole(tab.id)}
                  className={cn(
                    'px-2 py-0.5 rounded text-[10px] font-medium transition-colors whitespace-nowrap flex items-center gap-1 cursor-pointer',
                    active
                      ? 'bg-zinc-800 text-zinc-100 border border-zinc-700 shadow-xs'
                      : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-855'
                  )}
                >
                  <span>{tab.label}</span>
                  <span
                    className={cn(
                      'text-[9px] px-1 rounded-full',
                      active ? 'bg-zinc-700 text-zinc-200' : 'text-zinc-500'
                    )}
                  >
                    {tab.count}
                  </span>
                </button>
              );
            })}
          </div>

          {/* Search Input */}
          <div className="relative">
            <input
              type="text"
              placeholder="Filter by name, email, or role..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full h-7 pl-7 pr-2 text-[11px] rounded bg-zinc-900 border border-zinc-800 text-zinc-200 placeholder:text-zinc-500 focus:outline-none focus:border-zinc-700"
            />
            <Search className="absolute left-2 top-2 h-3 w-3 text-zinc-500" />
          </div>

          {/* Scrollable Account Cards */}
          <div className="space-y-1.5 max-h-52 overflow-y-auto pr-1">
            {filteredAccounts.length === 0 ? (
              <div className="py-4 text-center text-[11px] text-zinc-500">
                No accounts match your search.
              </div>
            ) : (
              filteredAccounts.map((account) => {
                const isSelected = selectedEmail === account.email;
                const isPending = account.status === 'PENDING';
                const isSubmitting =
                  isLoggingIn && activeSubmittingEmail === account.email;

                return (
                  <div
                    key={account.email}
                    onClick={() => handleAutoFill(account)}
                    className={cn(
                      'group flex items-center justify-between p-2 rounded-md border transition-all cursor-pointer text-left',
                      isSelected
                        ? 'border-indigo-500/70 bg-zinc-850/90 ring-1 ring-indigo-500/30'
                        : 'border-zinc-800/80 bg-zinc-900/50 hover:bg-zinc-850 hover:border-zinc-700'
                    )}
                  >
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <Avatar className="h-7 w-7 rounded-full border border-zinc-700/80 shrink-0 overflow-hidden">
                        {account.avatar ? (
                          <AvatarImage
                            src={account.avatar}
                            alt={account.name}
                            className="h-full w-full object-cover"
                          />
                        ) : null}
                        <AvatarFallback className="bg-zinc-800 text-zinc-200 text-[10px] font-semibold">
                          {getInitials(account.name)}
                        </AvatarFallback>
                      </Avatar>

                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-1.5 flex-wrap">
                          <span className="text-[11px] font-semibold text-zinc-100 group-hover:text-white truncate">
                            {account.name}
                          </span>
                          {getRoleBadge(account.role)}
                          {isPending && (
                            <span className="inline-flex items-center gap-0.5 rounded px-1 py-0 text-[8px] font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30">
                              <AlertCircle className="h-2 w-2" />
                              Pending
                            </span>
                          )}
                        </div>
                        <p className="text-[10px] text-zinc-400 truncate">
                          {account.email} &bull;{' '}
                          <span className="text-zinc-500">{account.title}</span>
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-1 shrink-0 ml-2">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAutoFill(account);
                        }}
                        disabled={isLoggingIn}
                        className="px-2 py-1 text-[10px] font-medium rounded bg-zinc-800 hover:bg-zinc-750 text-zinc-300 transition-colors cursor-pointer"
                        title="Auto-fill email and password"
                      >
                        Fill
                      </button>

                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleOneClickLogin(account);
                        }}
                        disabled={isLoggingIn}
                        className={cn(
                          'flex items-center gap-1 px-2 py-1 text-[10px] font-medium rounded text-white transition-colors cursor-pointer',
                          isPending
                            ? 'bg-zinc-800 hover:bg-zinc-700 text-zinc-400'
                            : 'bg-emerald-600 hover:bg-emerald-500'
                        )}
                        title={
                          isPending
                            ? 'Account is pending activation'
                            : '1-Click sign in as this user'
                        }
                      >
                        {isSubmitting ? (
                          <Loader2 className="h-2.5 w-2.5 animate-spin" />
                        ) : (
                          <Zap className="h-2.5 w-2.5 fill-current" />
                        )}
                        <span>Login</span>
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}

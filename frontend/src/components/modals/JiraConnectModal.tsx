'use client';

import * as React from 'react';
import { useState } from 'react';
import { ExternalLink, Check, Loader2, ArrowRight, ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  useConnectJiraWorkspaceMutation,
  useGetJiraProjectsQuery,
  useImportJiraProjectsMutation,
} from '@/lib/store/api/integrationsApi';

interface JiraConnectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  workspaceId: string;
}

export function JiraConnectModal({
  open,
  onOpenChange,
  workspaceId,
}: JiraConnectModalProps) {
  const [step, setStep] = useState<1 | 2>(1);

  // Form State
  const [jiraSiteUrl, setJiraSiteUrl] = useState('');
  const [jiraEmail, setJiraEmail] = useState('');
  const [jiraApiToken, setJiraApiToken] = useState('');
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  // Mutations & Queries
  const [connectJira, { isLoading: isConnecting }] = useConnectJiraWorkspaceMutation();
  const { data: remoteProjects = [], isLoading: isLoadingProjects } = useGetJiraProjectsQuery(
    workspaceId,
    { skip: step !== 2 || !workspaceId }
  );
  const [importProjects, { isLoading: isImporting }] = useImportJiraProjectsMutation();

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!jiraSiteUrl || !jiraEmail || !jiraApiToken) {
      toast.error('Please fill in all required Jira credentials');
      return;
    }

    try {
      await connectJira({
        workspaceId,
        data: {
          jiraSiteUrl: jiraSiteUrl.trim(),
          jiraEmail: jiraEmail.trim(),
          jiraApiToken: jiraApiToken.trim(),
        },
      }).unwrap();

      toast.success('Connected to Jira Cloud');
      setStep(2);
    } catch (err: unknown) {
      const errMsg =
        (err as { data?: { message?: string } })?.data?.message ||
        'Failed to connect to Jira. Check your site URL and API token.';
      toast.error(errMsg);
    }
  };

  const handleToggleProject = (key: string) => {
    setSelectedProjects((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key]
    );
  };

  const handleImport = async () => {
    if (selectedProjects.length === 0) {
      toast.info('Select at least one Jira project to import');
      return;
    }

    try {
      const result = await importProjects({
        workspaceId,
        projects: selectedProjects.map((key) => ({ key })),
      }).unwrap();

      toast.success(`Successfully imported ${result.imported || selectedProjects.length} project(s)`);
      onOpenChange(false);
      setStep(1);
      setSelectedProjects([]);
    } catch {
      toast.error('Failed to import Jira projects');
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-lg rounded-md p-6">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
              <span>Connect Jira Cloud</span>
              <Badge variant="outline" className="text-[10px] font-mono">
                Step {step} of 2
              </Badge>
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-zinc-400">
            {step === 1
              ? 'Enter your Atlassian Cloud credentials with an API token to link Jira.'
              : 'Select the Jira projects you wish to import and sync with Orbito.'}
          </DialogDescription>
        </DialogHeader>

        {step === 1 ? (
          <form onSubmit={handleConnect} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="jiraUrl" className="text-xs text-zinc-300">
                Atlassian Site URL
              </Label>
              <Input
                id="jiraUrl"
                placeholder="https://your-company.atlassian.net"
                value={jiraSiteUrl}
                onChange={(e) => setJiraSiteUrl(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-xs"
              />
              <p className="text-[11px] text-zinc-500">
                Must be an Atlassian Cloud domain (SSRF protected).
              </p>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="jiraEmail" className="text-xs text-zinc-300">
                Atlassian Account Email
              </Label>
              <Input
                id="jiraEmail"
                type="email"
                placeholder="developer@company.com"
                value={jiraEmail}
                onChange={(e) => setJiraEmail(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-xs"
              />
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label htmlFor="jiraToken" className="text-xs text-zinc-300">
                  Jira API Token
                </Label>
                <a
                  href="https://id.atlassian.com/manage-profile/security/api-tokens"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[10px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 transition-colors"
                >
                  <span>Generate Token</span>
                  <ExternalLink className="h-2.5 w-2.5" />
                </a>
              </div>
              <Input
                id="jiraToken"
                type="password"
                placeholder="••••••••••••••••••••••••"
                value={jiraApiToken}
                onChange={(e) => setJiraApiToken(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-xs font-mono"
              />
            </div>

            <DialogFooter className="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800/80">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => onOpenChange(false)}
                disabled={isConnecting}
                className="text-xs"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="sm"
                disabled={isConnecting}
                className="text-xs gap-1.5"
              >
                {isConnecting ? (
                  <>
                    <Loader2 className="h-3 w-3 animate-spin" />
                    <span>Verifying...</span>
                  </>
                ) : (
                  <>
                    <span>Next: Select Projects</span>
                    <ArrowRight className="h-3 w-3" />
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="space-y-4 pt-2">
            <div className="max-h-60 overflow-y-auto divide-y divide-zinc-800/60 border border-zinc-800/80 rounded-md bg-zinc-950/60">
              {isLoadingProjects ? (
                <div className="py-8 text-center text-xs text-zinc-500 flex items-center justify-center gap-2">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Discovering remote Jira projects...</span>
                </div>
              ) : remoteProjects.length === 0 ? (
                <div className="py-8 text-center text-xs text-zinc-500">
                  No projects found in this Jira site.
                </div>
              ) : (
                remoteProjects.map((p) => {
                  const isSelected = selectedProjects.includes(p.key);
                  return (
                    <div
                      key={p.id || p.key}
                      onClick={() => handleToggleProject(p.key)}
                      className={`p-3 flex items-center justify-between cursor-pointer transition-colors hover:bg-zinc-850/60 ${
                        isSelected ? 'bg-zinc-850/80' : ''
                      }`}
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div
                          className={`h-4 w-4 rounded-sm border flex items-center justify-center transition-colors ${
                            isSelected
                              ? 'bg-zinc-100 border-white text-zinc-900'
                              : 'border-zinc-700 bg-zinc-900'
                          }`}
                        >
                          {isSelected && <Check className="h-3 w-3 stroke-[3]" />}
                        </div>
                        <div className="min-w-0">
                          <p className="text-xs font-medium text-zinc-200 truncate">
                            {p.name}
                          </p>
                          <span className="text-[10px] font-mono text-zinc-500">
                            Key: {p.key}
                          </span>
                        </div>
                      </div>
                      <Badge variant="outline" className="text-[10px]">
                        {p.projectTypeKey || 'software'}
                      </Badge>
                    </div>
                  );
                })
              )}
            </div>

            <DialogFooter className="flex items-center justify-between gap-2 pt-4 border-t border-zinc-800/80">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setStep(1)}
                disabled={isImporting}
                className="text-xs gap-1"
              >
                <ArrowLeft className="h-3 w-3" />
                <span>Back</span>
              </Button>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => onOpenChange(false)}
                  disabled={isImporting}
                  className="text-xs"
                >
                  Close
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  size="sm"
                  onClick={handleImport}
                  disabled={isImporting || selectedProjects.length === 0}
                  className="text-xs gap-1.5"
                >
                  {isImporting ? (
                    <>
                      <Loader2 className="h-3 w-3 animate-spin" />
                      <span>Importing...</span>
                    </>
                  ) : (
                    <span>Import {selectedProjects.length} Project(s)</span>
                  )}
                </Button>
              </div>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}


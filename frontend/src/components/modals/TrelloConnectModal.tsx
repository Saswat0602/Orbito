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
  useConnectTrelloWorkspaceMutation,
  useGetTrelloBoardsQuery,
  useImportTrelloBoardsMutation,
} from '@/lib/store/api/integrationsApi';

interface TrelloConnectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  workspaceId: string;
}

export function TrelloConnectModal({
  open,
  onOpenChange,
  workspaceId,
}: TrelloConnectModalProps) {
  const [step, setStep] = useState<1 | 2>(1);

  // Form State
  const [trelloApiKey, setTrelloApiKey] = useState('');
  const [trelloToken, setTrelloToken] = useState('');
  const [selectedBoards, setSelectedBoards] = useState<string[]>([]);

  // Mutations & Queries
  const [connectTrello, { isLoading: isConnecting }] = useConnectTrelloWorkspaceMutation();
  const { data: remoteBoards = [], isLoading: isLoadingBoards } = useGetTrelloBoardsQuery(
    workspaceId,
    { skip: step !== 2 || !workspaceId }
  );
  const [importBoards, { isLoading: isImporting }] = useImportTrelloBoardsMutation();

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!trelloApiKey || !trelloToken) {
      toast.error('Please enter both your Trello API Key and Token');
      return;
    }

    try {
      await connectTrello({
        workspaceId,
        data: {
          trelloApiKey: trelloApiKey.trim(),
          trelloToken: trelloToken.trim(),
        },
      }).unwrap();

      toast.success('Connected to Trello');
      setStep(2);
    } catch (err: unknown) {
      const errMsg =
        (err as { data?: { message?: string } })?.data?.message ||
        'Failed to connect to Trello. Please check your API Key and Token.';
      toast.error(errMsg);
    }
  };

  const handleToggleBoard = (id: string) => {
    setSelectedBoards((prev) =>
      prev.includes(id) ? prev.filter((b) => b !== id) : [...prev, id]
    );
  };

  const handleImport = async () => {
    if (selectedBoards.length === 0) {
      toast.info('Select at least one Trello board to import');
      return;
    }

    try {
      const result = await importBoards({
        workspaceId,
        boardIds: selectedBoards,
      }).unwrap();

      toast.success(`Successfully imported ${result.imported || selectedBoards.length} board(s)`);
      onOpenChange(false);
      setStep(1);
      setSelectedBoards([]);
    } catch {
      toast.error('Failed to import Trello boards');
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-lg rounded-md p-6">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
              <span>Connect Trello</span>
              <Badge variant="outline" className="text-[10px] font-mono">
                Step {step} of 2
              </Badge>
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-zinc-400">
            {step === 1
              ? 'Enter your Trello Developer API Key and Token to link your account.'
              : 'Select the Trello boards you wish to import as Orbito projects.'}
          </DialogDescription>
        </DialogHeader>

        {step === 1 ? (
          <form onSubmit={handleConnect} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label htmlFor="trelloKey" className="text-xs text-zinc-300">
                  Trello API Key
                </Label>
                <a
                  href="https://trello.com/power-ups/admin"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[10px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 transition-colors"
                >
                  <span>Get API Key</span>
                  <ExternalLink className="h-2.5 w-2.5" />
                </a>
              </div>
              <Input
                id="trelloKey"
                placeholder="Paste your 32-character API key"
                value={trelloApiKey}
                onChange={(e) => setTrelloApiKey(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-xs font-mono"
              />
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label htmlFor="trelloToken" className="text-xs text-zinc-300">
                  Trello Token
                </Label>
                <a
                  href="https://trello.com/1/authorize?expiration=never&name=Orbito&scope=read,write&response_type=token&key=YOUR_KEY"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[10px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 transition-colors"
                >
                  <span>Authorize Token</span>
                  <ExternalLink className="h-2.5 w-2.5" />
                </a>
              </div>
              <Input
                id="trelloToken"
                type="password"
                placeholder="••••••••••••••••••••••••"
                value={trelloToken}
                onChange={(e) => setTrelloToken(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-xs font-mono"
              />
              <p className="text-[11px] text-zinc-500">
                Grant read/write access so Orbito can synchronize board lists.
              </p>
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
                    <span>Connecting...</span>
                  </>
                ) : (
                  <>
                    <span>Next: Select Boards</span>
                    <ArrowRight className="h-3 w-3" />
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="space-y-4 pt-2">
            <div className="max-h-60 overflow-y-auto divide-y divide-zinc-800/60 border border-zinc-800/80 rounded-md bg-zinc-950/60">
              {isLoadingBoards ? (
                <div className="py-8 text-center text-xs text-zinc-500 flex items-center justify-center gap-2">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Discovering available Trello boards...</span>
                </div>
              ) : remoteBoards.length === 0 ? (
                <div className="py-8 text-center text-xs text-zinc-500">
                  No open boards found in your Trello account.
                </div>
              ) : (
                remoteBoards.map((b) => {
                  const isSelected = selectedBoards.includes(b.id);
                  return (
                    <div
                      key={b.id}
                      onClick={() => handleToggleBoard(b.id)}
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
                            {b.name}
                          </p>
                          {b.desc && (
                            <p className="text-[11px] text-zinc-500 truncate">
                              {b.desc}
                            </p>
                          )}
                        </div>
                      </div>
                      <Badge variant="outline" className="text-[10px]">
                        Board
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
                  disabled={isImporting || selectedBoards.length === 0}
                  className="text-xs gap-1.5"
                >
                  {isImporting ? (
                    <>
                      <Loader2 className="h-3 w-3 animate-spin" />
                      <span>Importing...</span>
                    </>
                  ) : (
                    <span>Import {selectedBoards.length} Board(s)</span>
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


'use client';

import React, { useState, useMemo } from 'react';
import {
  Share2,
  Copy,
  Check,
  Trash2,
  Clock,
  ExternalLink,
  ShieldAlert,
  Loader2,
  Calendar,
} from 'lucide-react';
import { toast } from 'sonner';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';
import {
  useCreateTaskShareMutation,
  useGetTaskSharesQuery,
  useRevokeTaskShareMutation,
} from '@/lib/store/api/tasksApi';

interface ShareTaskDialogProps {
  isOpen: boolean;
  onClose: () => void;
  taskId: string;
  taskTitle: string;
}

export const ShareTaskDialog: React.FC<ShareTaskDialogProps> = ({
  isOpen,
  onClose,
  taskId,
  taskTitle,
}) => {
  const [expiresInDays, setExpiresInDays] = useState<number>(7);
  const [copiedToken, setCopiedToken] = useState<string | null>(null);
  const [shareToRevoke, setShareToRevoke] = useState<string | null>(null);

  // Queries & Mutations
  const { data: shares = [], isLoading: isSharesLoading } = useGetTaskSharesQuery(
    taskId,
    { skip: !isOpen || !taskId }
  );

  const [currentTime] = useState(() => Date.now());

  const [createTaskShare, { isLoading: isCreating }] = useCreateTaskShareMutation();
  const [revokeTaskShare, { isLoading: isRevoking }] = useRevokeTaskShareMutation();

  const handleCopyLink = async (shareUrl: string, token: string) => {
    try {
      const fullUrl =
        shareUrl.startsWith('http')
          ? shareUrl
          : `${window.location.origin}${shareUrl.startsWith('/') ? '' : '/'}${shareUrl}`;

      await navigator.clipboard.writeText(fullUrl);
      setCopiedToken(token);
      toast.success('Public task link copied to clipboard');
      setTimeout(() => setCopiedToken(null), 2000);
    } catch {
      toast.error('Failed to copy link to clipboard');
    }
  };

  const handleGenerateLink = async () => {
    if (!taskId) return;

    try {
      const res = await createTaskShare({
        taskId,
        expiresInDays,
      }).unwrap();

      toast.success('Public share link generated');
      if (res.shareUrl) {
        handleCopyLink(res.shareUrl, res.token);
      }
    } catch {
      toast.error('Failed to create public share link');
    }
  };

  const handleConfirmRevoke = async () => {
    if (!shareToRevoke) return;
    try {
      await revokeTaskShare({ shareId: shareToRevoke, taskId }).unwrap();
      toast.success('Share link revoked successfully');
      setShareToRevoke(null);
    } catch {
      toast.error('Failed to revoke share link');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Share Task Publicly"
      description={`Generate secure, unauthenticated read-only links for "${taskTitle}". Anyone with this link can view task details without logging in.`}
    >
      <div className="space-y-5 pt-2">
        {/* Generate Link Control Card */}
        <div className="rounded-md border border-zinc-800 bg-zinc-950/80 p-4 space-y-3.5">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-1">
              <Label className="text-xs font-semibold text-zinc-200">
                Link Expiration
              </Label>
              <p className="text-[11px] text-zinc-500">
                Choose how long the public link remains valid before expiring.
              </p>
            </div>

            <Select
              value={String(expiresInDays)}
              onValueChange={(val) => setExpiresInDays(Number(val))}
              disabled={isCreating}
            >
              <SelectTrigger className="w-36 bg-zinc-900 border-zinc-700 text-zinc-200 h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                <SelectItem value="1" className="text-xs">1 Day</SelectItem>
                <SelectItem value="3" className="text-xs">3 Days</SelectItem>
                <SelectItem value="7" className="text-xs">7 Days (Default)</SelectItem>
                <SelectItem value="14" className="text-xs">14 Days</SelectItem>
                <SelectItem value="30" className="text-xs">30 Days</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Button
            type="button"
            variant="primary"
            size="sm"
            onClick={handleGenerateLink}
            disabled={isCreating}
            className="w-full gap-1.5 text-xs h-8"
          >
            {isCreating ? (
              <>
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                <span>Generating Link...</span>
              </>
            ) : (
              <>
                <Share2 className="h-3.5 w-3.5" />
                <span>Create & Copy Share Link</span>
              </>
            )}
          </Button>
        </div>

        {/* Active Shares Section */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-xs text-zinc-400">
            <span className="font-semibold text-zinc-300">Active Public Links</span>
            <span>{shares.length} Active</span>
          </div>

          {isSharesLoading ? (
            <div className="py-4 text-center text-xs text-zinc-500">
              Loading active links...
            </div>
          ) : shares.length === 0 ? (
            <div className="rounded-md border border-zinc-850 bg-zinc-950/40 p-4 text-center text-xs text-zinc-500 italic">
              No active share links. Generate a link above to share this task.
            </div>
          ) : (
            <div className="max-h-56 overflow-y-auto space-y-2">
              {shares.map((share) => {
                const expiresDate = new Date(share.expiresAt);
                const isExpired = expiresDate.getTime() < currentTime;
                const isCopied = copiedToken === share.token;

                return (
                  <div
                    key={share.id}
                    className="rounded-md border border-zinc-800 bg-zinc-900/60 p-3 flex items-center justify-between gap-3 text-xs"
                  >
                    <div className="min-w-0 space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-[11px] text-zinc-300 truncate">
                          token: {share.token.slice(0, 12)}...
                        </span>
                        {isExpired ? (
                          <Badge variant="outline" className="text-[9px] text-red-400 border-red-800">
                            Expired
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="text-[9px] text-emerald-400 border-emerald-800">
                            Active
                          </Badge>
                        )}
                      </div>

                      <div className="flex items-center gap-2 text-[10px] text-zinc-500">
                        <Clock className="h-3 w-3" />
                        <span>
                          Expires: {expiresDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-1.5 shrink-0">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => handleCopyLink(share.shareUrl, share.token)}
                        className="h-7 px-2 gap-1 text-[11px] border-zinc-700 bg-zinc-800 text-zinc-200"
                        title="Copy Link"
                      >
                        {isCopied ? (
                          <>
                            <Check className="h-3 w-3 text-emerald-400" />
                            <span>Copied</span>
                          </>
                        ) : (
                          <>
                            <Copy className="h-3 w-3" />
                            <span>Copy</span>
                          </>
                        )}
                      </Button>

                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => setShareToRevoke(share.id)}
                        disabled={isRevoking}
                        className="h-7 w-7 p-0 text-zinc-500 hover:text-red-400"
                        title="Revoke Share Link"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Security Notice */}
        <div className="rounded-md border border-zinc-800/80 bg-zinc-950/40 p-3 flex items-start gap-2.5 text-[11px] text-zinc-400">
          <ShieldAlert className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
          <span>
            Shared links provide read-only access. Sensitive comments, audit trails, and internal workspace directories are strictly excluded from the public view.
          </span>
        </div>
      </div>

      {/* Revoke Confirmation Modal */}
      <ConfirmationModal
        open={!!shareToRevoke}
        onOpenChange={(open) => !open && setShareToRevoke(null)}
        title="Revoke Public Share Link"
        description="Are you sure you want to revoke this public link? Anyone attempting to visit this link will immediately receive a 404 access expired page."
        confirmLabel="Revoke Link"
        variant="destructive"
        isLoading={isRevoking}
        onConfirm={handleConfirmRevoke}
      />
    </Modal>
  );
};
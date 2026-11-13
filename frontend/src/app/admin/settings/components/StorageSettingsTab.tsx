'use client';

import * as React from 'react';
import { HardDrive, Server, Eye, EyeOff, Save, CheckCircle2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface StorageSettingsTabProps {
  storageProvider: 'local' | 's3';
  setStorageProvider: (val: 'local' | 's3') => void;
  s3Bucket: string;
  setS3Bucket: (val: string) => void;
  s3Region: string;
  setS3Region: (val: string) => void;
  s3Endpoint: string;
  setS3Endpoint: (val: string) => void;
  s3AccessKey: string;
  setS3AccessKey: (val: string) => void;
  s3SecretKey: string;
  setS3SecretKey: (val: string) => void;
  showS3Secret: boolean;
  setShowS3Secret: (val: boolean) => void;
  maxUploadMb: string;
  setMaxUploadMb: (val: string) => void;
  onSave: () => void;
  isSaving: boolean;
}

export function StorageSettingsTab({
  storageProvider,
  setStorageProvider,
  s3Bucket,
  setS3Bucket,
  s3Region,
  setS3Region,
  s3Endpoint,
  setS3Endpoint,
  s3AccessKey,
  setS3AccessKey,
  s3SecretKey,
  setS3SecretKey,
  showS3Secret,
  setShowS3Secret,
  maxUploadMb,
  setMaxUploadMb,
  onSave,
  isSaving,
}: StorageSettingsTabProps) {
  return (
    <div className="space-y-6">
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
        {/* Row 1: Storage Driver Selector */}
        <div className="p-4 space-y-3">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">Asset Storage Engine</span>
            <p className="text-xs text-zinc-400">
              Where user avatars, task attachments, and project exports are persisted.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
            <div
              onClick={() => setStorageProvider('local')}
              className={`p-3 rounded-md border cursor-pointer transition-all flex items-start gap-3 ${
                storageProvider === 'local'
                  ? 'bg-zinc-850 border-zinc-600 text-zinc-100 shadow-xs'
                  : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:border-zinc-700 hover:text-zinc-300'
              }`}
            >
              <div className="p-1.5 rounded-md bg-zinc-900 border border-zinc-800 shrink-0 text-zinc-300">
                <HardDrive className="h-4 w-4" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-1.5 font-medium text-zinc-200 text-xs">
                  <span>Local Filesystem</span>
                  {storageProvider === 'local' && <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />}
                </div>
                <p className="text-[11px] text-zinc-400 leading-normal">
                  Stored directly on server disk volume under `/uploads`. Suitable for single-node deployments.
                </p>
              </div>
            </div>

            <div
              onClick={() => setStorageProvider('s3')}
              className={`p-3 rounded-md border cursor-pointer transition-all flex items-start gap-3 ${
                storageProvider === 's3'
                  ? 'bg-zinc-850 border-zinc-600 text-zinc-100 shadow-xs'
                  : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:border-zinc-700 hover:text-zinc-300'
              }`}
            >
              <div className="p-1.5 rounded-md bg-zinc-900 border border-zinc-800 shrink-0 text-zinc-300">
                <Server className="h-4 w-4" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-1.5 font-medium text-zinc-200 text-xs">
                  <span>S3-Compatible Cloud Storage</span>
                  {storageProvider === 's3' && <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />}
                </div>
                <p className="text-[11px] text-zinc-400 leading-normal">
                  AWS S3, Cloudflare R2, MinIO, or Wasabi object storage. Recommended for multi-region & scale.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Row 2: S3 Parameters (Conditional) */}
        {storageProvider === 's3' && (
          <div className="p-4 space-y-3">
            <span className="text-xs font-medium text-zinc-300">S3 Cloud Credentials & Endpoint</span>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-zinc-400">S3 Bucket Name</label>
                <Input
                  value={s3Bucket}
                  onChange={(e) => setS3Bucket(e.target.value)}
                  placeholder="orbito-assets-prod"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-zinc-400">AWS Region</label>
                <Input
                  value={s3Region}
                  onChange={(e) => setS3Region(e.target.value)}
                  placeholder="us-east-1"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                />
              </div>

              <div className="space-y-1.5 sm:col-span-2">
                <label className="text-[11px] font-medium text-zinc-400">
                  Custom Endpoint URL <span className="text-zinc-500">(Required for MinIO, Cloudflare R2, Wasabi)</span>
                </label>
                <Input
                  value={s3Endpoint}
                  onChange={(e) => setS3Endpoint(e.target.value)}
                  placeholder="https://<accountid>.r2.cloudflarestorage.com or https://minio.yourdomain.com"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-zinc-400">Access Key ID</label>
                <Input
                  value={s3AccessKey}
                  onChange={(e) => setS3AccessKey(e.target.value)}
                  placeholder="AKIAIOSFODNN7EXAMPLE"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[11px] font-medium text-zinc-400">Secret Access Key</label>
                <div className="relative">
                  <Input
                    type={showS3Secret ? 'text' : 'password'}
                    value={s3SecretKey}
                    onChange={(e) => setS3SecretKey(e.target.value)}
                    placeholder="••••••••••••••••••••••••"
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200 pr-8"
                  />
                  <button
                    type="button"
                    onClick={() => setShowS3Secret(!showS3Secret)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
                  >
                    {showS3Secret ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Row 3: Max Upload Size */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <span className="text-sm font-medium text-zinc-200">Max Upload File Size (MB)</span>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Limit maximum file size per individual attachment across tasks, comments, and project briefs.
            </p>
          </div>
          <div className="w-full sm:w-48 shrink-0">
            <Input
              value={maxUploadMb}
              onChange={(e) => setMaxUploadMb(e.target.value)}
              placeholder="50"
              className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
            />
          </div>
        </div>
      </div>

      {/* Save Button for Storage Tab */}
      <div className="flex justify-end">
        <Button
          type="button"
          size="sm"
          onClick={onSave}
          disabled={isSaving}
          className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
        >
          <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
          Save Storage Settings
        </Button>
      </div>
    </div>
  );
}

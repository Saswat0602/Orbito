'use client';

import * as React from 'react';
import { Globe, Save, ShieldCheck, AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface GeneralSettingsTabProps {
  platformName: string;
  setPlatformName: (val: string) => void;
  supportEmail: string;
  setSupportEmail: (val: string) => void;
  defaultTimezone: string;
  setDefaultTimezone: (val: string) => void;
  isMaintenanceMode: boolean;
  onToggleMaintenance: () => void;
  isPublicRegEnabled: boolean;
  onToggleRegistration: () => void;
  onSave: () => void;
  isSaving: boolean;
}

export function GeneralSettingsTab({
  platformName,
  setPlatformName,
  supportEmail,
  setSupportEmail,
  defaultTimezone,
  setDefaultTimezone,
  isMaintenanceMode,
  onToggleMaintenance,
  isPublicRegEnabled,
  onToggleRegistration,
  onSave,
  isSaving,
}: GeneralSettingsTabProps) {
  return (
    <div className="space-y-6">
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
        {/* Row 1: Public Registration Gate */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">Public User Registration</span>
              <span
                className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${
                  isPublicRegEnabled
                    ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60'
                    : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                }`}
              >
                {isPublicRegEnabled ? 'Open (/register)' : 'Restricted (Invite Only)'}
              </span>
            </div>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Allow external visitors to self-register their own organizations. When restricted, onboarding requires an explicit workspace or tenant invite token.
            </p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Switch
              checked={isPublicRegEnabled}
              onCheckedChange={onToggleRegistration}
              disabled={isSaving}
              aria-label="Toggle Public Registration"
            />
          </div>
        </div>

        {/* Row 2: Maintenance Mode */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">Maintenance Mode Gate</span>
              <span
                className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${
                  isMaintenanceMode
                    ? 'bg-amber-950/80 text-amber-400 border border-amber-800/60'
                    : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                }`}
              >
                {isMaintenanceMode ? 'Maintenance Active' : 'Operational'}
              </span>
            </div>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Temporarily pause public and standard member requests during infrastructure maintenance. Platform Super Admins retain full access.
            </p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Switch
              checked={isMaintenanceMode}
              onCheckedChange={onToggleMaintenance}
              disabled={isSaving}
              aria-label="Toggle Maintenance Mode"
            />
          </div>
        </div>

        {/* Row 3: Platform Instance Name */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <span className="text-sm font-medium text-zinc-200">Platform Instance Name</span>
            <p className="text-xs text-zinc-400 leading-relaxed">
              The name displayed on browser tabs, notification emails, and top navigation shells.
            </p>
          </div>
          <div className="w-full sm:w-72 shrink-0">
            <Input
              value={platformName}
              onChange={(e) => setPlatformName(e.target.value)}
              placeholder="Orbito Enterprise"
              className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
            />
          </div>
        </div>

        {/* Row 4: Support Contact Email */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <span className="text-sm font-medium text-zinc-200">System Support Email</span>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Primary administrative contact displayed to users encountering permission or billing gates.
            </p>
          </div>
          <div className="w-full sm:w-72 shrink-0">
            <Input
              value={supportEmail}
              onChange={(e) => setSupportEmail(e.target.value)}
              placeholder="support@orbito.io"
              className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
            />
          </div>
        </div>

        {/* Row 5: Reference Timezone */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <span className="text-sm font-medium text-zinc-200">Default Reference Timezone</span>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Default time reference for scheduled automated jobs, sprint boundaries, and daily activity digests.
            </p>
          </div>
          <div className="w-full sm:w-72 shrink-0">
            <Select
              value={defaultTimezone}
              onValueChange={(val) => setDefaultTimezone(val)}
            >
              <SelectTrigger className="w-full h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                <SelectItem value="UTC" className="text-xs">UTC (Universal Coordinated)</SelectItem>
                <SelectItem value="America/New_York" className="text-xs">Eastern Time (US / New York)</SelectItem>
                <SelectItem value="America/Los_Angeles" className="text-xs">Pacific Time (US / Los Angeles)</SelectItem>
                <SelectItem value="Europe/London" className="text-xs">London (GMT / BST)</SelectItem>
                <SelectItem value="Asia/Tokyo" className="text-xs">Tokyo (JST)</SelectItem>
                <SelectItem value="Asia/Kolkata" className="text-xs">India Standard Time (IST)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Save Button for General Tab */}
      <div className="flex justify-end">
        <Button
          type="button"
          size="sm"
          onClick={onSave}
          disabled={isSaving}
          className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
        >
          <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
          Save General Settings
        </Button>
      </div>
    </div>
  );
}

'use client';

import * as React from 'react';
import { useState } from 'react';
import { FolderGit2, Plus } from 'lucide-react';
import { toast } from 'sonner';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { useCreateProjectMutation } from '@/lib/store/api/projectsApi';

interface CreateProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId: string;
}

export function CreateProjectModal({
  isOpen,
  onClose,
  workspaceId,
}: CreateProjectModalProps) {
  const [createProject, { isLoading }] = useCreateProjectMutation();

  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [description, setDescription] = useState('');

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setName(val);
    // Auto-generate slug from name
    const generated = val
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
    setSlug(generated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !workspaceId) return;

    try {
      await createProject({
        name: name.trim(),
        slug: slug.trim() || undefined,
        description: description.trim() || undefined,
        workspaceId,
      }).unwrap();

      toast.success('Project created successfully');
      setName('');
      setSlug('');
      setDescription('');
      onClose();
    } catch {
      toast.error('Failed to create project');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New Project"
      description="Create a project to organize tasks, backlogs, sprints, and roadmaps."
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        <div className="space-y-1.5">
          <Label htmlFor="proj-name" className="text-xs text-zinc-300">
            Project Name
          </Label>
          <Input
            id="proj-name"
            value={name}
            onChange={handleNameChange}
            placeholder="e.g. Infrastructure Platform"
            required
            className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="proj-slug" className="text-xs text-zinc-300">
            Project Identifier / Slug
          </Label>
          <Input
            id="proj-slug"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="e.g. infra-platform"
            className="font-mono bg-zinc-950 border-zinc-800 text-xs text-zinc-300"
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="proj-desc" className="text-xs text-zinc-300">
            Description
          </Label>
          <Textarea
            id="proj-desc"
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Scope, objectives, and deliverables..."
            className="bg-zinc-950 border-zinc-800 text-xs text-zinc-200 resize-none"
          />
        </div>

        <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClose}
            className="text-xs"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            isLoading={isLoading}
            className="text-xs gap-1.5"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Create Project</span>
          </Button>
        </div>
      </form>
    </Modal>
  );
}


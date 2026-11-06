'use client';

import React, { useState, useEffect } from 'react';
import { Task, TaskStatus, TaskPriority, TaskType } from '@/types';
import { WEEKS, STATUSES } from '@/lib/utils/constants';
import { Plus, X } from 'lucide-react';
import { CustomSelect } from '../ui/CustomSelect';

interface TaskModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (task: Partial<Task>) => void;
    initialStatus?: TaskStatus;
    initialWeek?: number;
}

const priorityOptions: { value: TaskPriority; label: string }[] = [
    { value: 'low', label: 'Low Priority' },
    { value: 'medium', label: 'Medium Priority' },
    { value: 'high', label: 'High Priority' },
];

const typeOptions: { value: TaskType; label: string }[] = [
    { value: 'feature', label: 'Feature' },
    { value: 'bug', label: 'Bug' },
    { value: 'improvement', label: 'Improvement' },
    { value: 'research', label: 'Research' },
    { value: 'meeting', label: 'Meeting' },
    { value: 'other', label: 'Other' },
];

export const TaskModal: React.FC<TaskModalProps> = ({
    isOpen,
    onClose,
    onSubmit,
    initialStatus = 'todo',
    initialWeek = 1,
}) => {
    const [formData, setFormData] = useState<Partial<Task>>({
        text: '',
        description: '',
        status: initialStatus,
        priority: 'medium',
        type: 'other',
        week: initialWeek,
        tags: [],
    });

    const [tagInput, setTagInput] = useState('');

    useEffect(() => {
        // Reset form when modal opens
        setFormData({
            text: '',
            description: '',
            status: initialStatus,
            priority: 'medium',
            type: 'other',
            week: initialWeek,
            tags: [],
        });
    }, [isOpen, initialStatus, initialWeek]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.text?.trim()) return;
        onSubmit(formData);
        onClose();
    };

    const handleAddTag = () => {
        if (tagInput.trim() && !formData.tags?.includes(tagInput.trim())) {
            setFormData({
                ...formData,
                tags: [...(formData.tags || []), tagInput.trim()],
            });
            setTagInput('');
        }
    };

    const handleRemoveTag = (tag: string) => {
        setFormData({
            ...formData,
            tags: formData.tags?.filter((t) => t !== tag) || [],
        });
    };

    if (!isOpen) return null;

    const weekOptions = WEEKS.map((w) => ({ value: w.id.toString(), label: w.label }));
    const statusOptions = STATUSES.map((s) => ({ value: s.id, label: s.label }));

    return (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-fade-in">
            <div className="bg-[rgb(var(--color-bg-primary))] rounded-3xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl animate-zoom-in border border-[rgb(var(--color-border-primary))] scrollbar-thin">
                <form onSubmit={handleSubmit} className="p-6 md:p-8 space-y-6">
                    {/* Header */}
                    <div className="flex justify-between items-start pb-4 border-b border-[rgb(var(--color-border-secondary))]">
                        <div className="flex-1">
                            <h3 className="text-2xl font-bold">New Task</h3>
                            <p className="text-sm text-[rgb(var(--color-text-tertiary))] mt-1">
                                Create a new task for your workflow
                            </p>
                        </div>
                        <button
                            type="button"
                            onClick={onClose}
                            className="p-2 hover:bg-[rgb(var(--color-bg-tertiary))] rounded-full transition-all ml-2"
                        >
                            <X size={20} className="text-[rgb(var(--color-text-tertiary))]" />
                        </button>
                    </div>

                    {/* Form Fields */}
                    <div className="space-y-5">
                        {/* Task Title */}
                        <div>
                            <label className="text-[10px] font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-[0.2em] mb-3 block">
                                Task Title
                            </label>
                            <input
                                autoFocus
                                type="text"
                                placeholder="e.g., Conduct user research"
                                className="w-full px-4 py-3 rounded-xl bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] transition-all font-semibold text-sm"
                                value={formData.text}
                                onChange={(e) => setFormData({ ...formData, text: e.target.value })}
                                required
                            />
                        </div>

                        {/* Description */}
                        <div>
                            <label className="text-[10px] font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-[0.2em] mb-3 block">
                                Description (Optional)
                            </label>
                            <textarea
                                placeholder="Add more details about this task..."
                                className="w-full px-4 py-3 rounded-xl bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] transition-all font-medium resize-none text-sm"
                                rows={3}
                                value={formData.description}
                                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            />
                        </div>

                        {/* Week and Priority */}
                        <div className="grid grid-cols-2 gap-4">
                            <CustomSelect
                                label="Week"
                                value={formData.week?.toString() || '1'}
                                onValueChange={(value) => setFormData({ ...formData, week: parseInt(value) })}
                                options={weekOptions}
                                placeholder="Select week"
                            />

                            <CustomSelect
                                label="Priority"
                                value={formData.priority || 'medium'}
                                onValueChange={(value) => setFormData({ ...formData, priority: value as TaskPriority })}
                                options={priorityOptions}
                                placeholder="Select priority"
                            />
                        </div>

                        {/* Status and Type */}
                        <div className="grid grid-cols-2 gap-4">
                            <CustomSelect
                                label="Status"
                                value={formData.status || 'todo'}
                                onValueChange={(value) => setFormData({ ...formData, status: value as TaskStatus })}
                                options={statusOptions}
                                placeholder="Select status"
                            />

                            <CustomSelect
                                label="Type"
                                value={formData.type || 'other'}
                                onValueChange={(value) => setFormData({ ...formData, type: value as TaskType })}
                                options={typeOptions}
                                placeholder="Select type"
                            />
                        </div>

                        {/* Tags */}
                        <div>
                            <label className="text-[10px] font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-[0.2em] mb-3 block">
                                Tags
                            </label>
                            <div className="flex gap-2 mb-3">
                                <input
                                    type="text"
                                    placeholder="Add a tag..."
                                    className="flex-1 px-4 py-2.5 rounded-xl bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] transition-all text-sm"
                                    value={tagInput}
                                    onChange={(e) => setTagInput(e.target.value)}
                                    onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddTag())}
                                />
                                <button
                                    type="button"
                                    onClick={handleAddTag}
                                    className="px-4 py-2.5 bg-[rgb(var(--color-primary))] text-white rounded-xl hover:bg-[rgb(var(--color-primary-hover))] transition-all"
                                >
                                    <Plus size={18} />
                                </button>
                            </div>
                            {formData.tags && formData.tags.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                    {formData.tags.map((tag) => (
                                        <span
                                            key={tag}
                                            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))] rounded-full text-xs font-bold"
                                        >
                                            {tag}
                                            <button
                                                type="button"
                                                onClick={() => handleRemoveTag(tag)}
                                                className="hover:text-[rgb(var(--color-error))]"
                                            >
                                                <X size={14} />
                                            </button>
                                        </span>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Submit Button */}
                    <button
                        type="submit"
                        className="w-full bg-[rgb(var(--color-primary))] text-white font-black py-4 rounded-2xl hover:bg-[rgb(var(--color-primary-hover))] transition-all shadow-xl uppercase tracking-widest text-sm"
                    >
                        Create Task
                    </button>
                </form>
            </div>
        </div>
    );
};
